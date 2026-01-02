package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	observabilityv1alpha1 "github.com/alexandrosst/observability-gateway-operator/api/v1alpha1"
)

// ObservabilityGatewayReconciler reconciles a ObservabilityGateway object
type ObservabilityGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=observability.x-k8s.io,resources=observabilitygateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.x-k8s.io,resources=observabilitygateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *ObservabilityGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	gw := &observabilityv1alpha1.ObservabilityGateway{}
	if err := r.Get(ctx, req.NamespacedName, gw); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Reconciling ObservabilityGateway", "name", gw.Name, "classes", len(gw.Spec.Classes))

	// Track current classes for deletion
	classNames := make(map[string]struct{})
	for _, class := range gw.Spec.Classes {
		classNames[class.Name] = struct{}{}
		if err := r.reconcileDeployment(ctx, gw, class); err != nil {
			log.Error(err, "Failed to reconcile Deployment", "class", class.Name)
			return ctrl.Result{}, err
		}
		if err := r.reconcileService(ctx, gw, class); err != nil {
			log.Error(err, "Failed to reconcile Service", "class", class.Name)
			return ctrl.Result{}, err
		}
	}

	// Delete removed classes
	if err := r.cleanupRemovedClasses(ctx, gw, classNames); err != nil {
		log.Error(err, "Failed to cleanup removed classes")
		return ctrl.Result{}, err
	}

	// TODO: Update status

	return ctrl.Result{}, nil
}

func (r *ObservabilityGatewayReconciler) reconcileDeployment(ctx context.Context, gw *observabilityv1alpha1.ObservabilityGateway, class observabilityv1alpha1.GatewayClass) error {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", gw.Name, class.Name),
			Namespace: gw.Namespace,
		},
	}

	log := log.FromContext(ctx)

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		dep.Labels = map[string]string{
			"app.kubernetes.io/part-of": gw.Name,
			"observability-class":       class.Name,
		}

		// Default ports if none specified
		ports := class.Ports
		if len(ports) == 0 {
			ports = []corev1.ContainerPort{
				{Name: "otlp-grpc", ContainerPort: 4317, Protocol: corev1.ProtocolTCP},
				{Name: "otlp-http", ContainerPort: 4318, Protocol: corev1.ProtocolTCP},
				{Name: "metrics", ContainerPort: 8888, Protocol: corev1.ProtocolTCP},
			}
		}

		// Resources: safe copy
		var resources corev1.ResourceRequirements
		if class.Resources != nil {
			resources = *class.Resources.DeepCopy()
		}

		// Affinity: safe
		var affinity *corev1.Affinity
		if class.Affinity != nil {
			affinity = class.Affinity.DeepCopy()
		}

		dep.Spec = appsv1.DeploymentSpec{
			Replicas: &class.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": dep.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": dep.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "otel-collector",
							Image: gw.Spec.Image,
							Args:  append(append([]string{"--config=/etc/otel/config.yaml"}, gw.Spec.ExtraArgs...), class.ExtraArgs...),
							Ports: ports,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "otel-config", MountPath: "/etc/otel"},
							},
							Resources: resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "otel-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: gw.Spec.ConfigConfigMap},
								},
							},
						},
					},
					NodeSelector: class.NodeSelector,
					Tolerations:  class.Tolerations,
					Affinity:     affinity,
				},
			},
		}

		return controllerutil.SetControllerReference(gw, dep, r.Scheme)
	})

	if err != nil {
		return err
	}

	log.Info("Deployment reconciled", "op", op, "name", dep.Name)
	return nil
}

func (r *ObservabilityGatewayReconciler) reconcileService(ctx context.Context, gw *observabilityv1alpha1.ObservabilityGateway, class observabilityv1alpha1.GatewayClass) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", gw.Name, class.Name),
			Namespace: gw.Namespace,
		},
	}

	log := log.FromContext(ctx)

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Labels = map[string]string{
			"app.kubernetes.io/part-of": gw.Name,
			"observability-class":       class.Name,
		}

		// Default ports if none specified
		containerPorts := class.Ports
		if len(containerPorts) == 0 {
			containerPorts = []corev1.ContainerPort{
				{Name: "otlp-grpc", ContainerPort: 4317, Protocol: corev1.ProtocolTCP},
				{Name: "otlp-http", ContainerPort: 4318, Protocol: corev1.ProtocolTCP},
				{Name: "metrics", ContainerPort: 8888, Protocol: corev1.ProtocolTCP},
			}
		}

		servicePorts := make([]corev1.ServicePort, len(containerPorts))
		for i, p := range containerPorts {
			servicePorts[i] = corev1.ServicePort{
				Name:       p.Name,
				Port:       p.ContainerPort,
				Protocol:   p.Protocol,
				TargetPort: intstr.FromInt(int(p.ContainerPort)),
			}
		}

		svc.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": svc.Name},
			Ports:    servicePorts,
			Type:     corev1.ServiceTypeClusterIP,
		}

		return controllerutil.SetControllerReference(gw, svc, r.Scheme)
	})

	if err != nil {
		return err
	}

	log.Info("Service reconciled", "op", op, "name", svc.Name)
	return nil
}

func (r *ObservabilityGatewayReconciler) cleanupRemovedClasses(ctx context.Context, gw *observabilityv1alpha1.ObservabilityGateway, currentClasses map[string]struct{}) error {
	// List all owned Deployments
	deployList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployList, client.InNamespace(gw.Namespace), client.MatchingLabels{"app.kubernetes.io/part-of": gw.Name}); err != nil {
		return err
	}
	for _, dep := range deployList.Items {
		className := dep.Labels["observability-class"]
		if _, exists := currentClasses[className]; !exists {
			if err := r.Delete(ctx, &dep); err != nil {
				return err
			}
		}
	}

	// List all owned Services
	svcList := &corev1.ServiceList{}
	if err := r.List(ctx, svcList, client.InNamespace(gw.Namespace), client.MatchingLabels{"app.kubernetes.io/part-of": gw.Name}); err != nil {
		return err
	}
	for _, svc := range svcList.Items {
		className := svc.Labels["observability-class"]
		if _, exists := currentClasses[className]; !exists {
			if err := r.Delete(ctx, &svc); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ObservabilityGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&observabilityv1alpha1.ObservabilityGateway{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
