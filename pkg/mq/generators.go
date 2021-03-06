/**
 * Copyright (c) 2020 xxxx . All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0 */

package mq

import (
	"fmt"
	"reflect"
	"strconv"
	
	"mosquitto-operator/pkg/apis/edge/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func headlessDomain(z * v1alpha1.MosquittoCluster)string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", headlessSvcName(z), z.GetNamespace())
}

func headlessSvcName(z * v1alpha1.MosquittoCluster)string {
	return fmt.Sprintf("%s-headless", z.GetName())
}

// MakeStatefulSet return a mosquitto stateful set from the zk spec
func MakeStatefulSet(z * v1alpha1.MosquittoCluster) * appsv1.StatefulSet {
	return & appsv1.StatefulSet {
		TypeMeta:metav1.TypeMeta {
			Kind:"StatefulSet", 
			APIVersion:"apps/v1", 
		}, 
		ObjectMeta:metav1.ObjectMeta {
			Name:z.GetName(), 
			Namespace:z.Namespace, 
			Labels:z.Spec.Labels, 
		}, 
		Spec:appsv1.StatefulSetSpec {
			ServiceName:headlessSvcName(z), 
			Replicas: & z.Spec.Replicas, 
			Selector: & metav1.LabelSelector {
				MatchLabels:map[string]string {
					"app":z.GetName(), 
				}, 
			}, 
			UpdateStrategy:appsv1.StatefulSetUpdateStrategy {
				Type:appsv1.RollingUpdateStatefulSetStrategyType, 
			}, 
			PodManagementPolicy:appsv1.OrderedReadyPodManagement, 
			Template:v1.PodTemplateSpec {
				ObjectMeta:metav1.ObjectMeta {
					GenerateName:z.GetName(), 
					Labels:map[string]string {
						"app":z.GetName(), 
						"kind":"MosquittoMember", 
					}, 
				}, 
				Spec:makeMqPodSpec(z), 
			}, 
			VolumeClaimTemplates:[]v1.PersistentVolumeClaim { {
					ObjectMeta:metav1.ObjectMeta {
						Name:"data", 
						Labels:map[string]string {"app":z.GetName()}, 
					}, 
					Spec:z.Spec.Persistence.PersistentVolumeClaimSpec, 
				},  
			}, 
		}, 
	}
}

func makeMqPodSpec(z * v1alpha1.MosquittoCluster)v1.PodSpec {
	mqContainer := v1.Container {
		Name:"mosquitto-broker", 
		Image:z.Spec.Image.ToString(), 
		Ports:z.Spec.Ports, 
		ImagePullPolicy:z.Spec.Image.PullPolicy, 
		//ReadinessProbe: & v1.Probe {
		//	InitialDelaySeconds:10, 
		//	TimeoutSeconds:10, 
			//Handler:v1.Handler {
			//	Exec: & v1.ExecAction {Command:[]string {"mosquittoReady.sh"}}, 
			//}, 
		//}, 
		//LivenessProbe: & v1.Probe {
		//	InitialDelaySeconds:10, 
		//	TimeoutSeconds:10, 
			//Handler:v1.Handler {
			//	Exec: & v1.ExecAction {Command:[]string {"mosquittoLive.sh"}}, 
			//}, 
		//}, 
		VolumeMounts:[]v1.VolumeMount { {Name:"data", MountPath:"/mosquitto/data"},  {Name:"conf", MountPath:"/mosquitto/config"}, 
		}, 
	
		//Lifecycle: & v1.Lifecycle {
		//	PreStop: & v1.Handler {
		//		Exec: & v1.ExecAction {
		//			Command:[]string {"mosquittoTeardown.sh"}, 
		//		}, 
		//	}, 
		//}, 
		//Command:[]string {"/usr/local/bin/mosquittoStart.sh"}, 
	}
	if z.Spec.Pod.Resources.Limits != nil || z.Spec.Pod.Resources.Requests != nil {
		mqContainer.Resources = z.Spec.Pod.Resources
	}
	mqContainer.Env = z.Spec.Pod.Env
	podSpec := v1.PodSpec {
		Containers:[]v1.Container {mqContainer}, 
		Affinity:z.Spec.Pod.Affinity, 
		Volumes:[]v1.Volume { {
				Name:"conf", 
				VolumeSource:v1.VolumeSource {
					ConfigMap: & v1.ConfigMapVolumeSource {
						LocalObjectReference:v1.LocalObjectReference {
							Name:z.ConfigMapName(), 
						}, 
					}, 
				}, 
			}, 
		}, 
		TerminationGracePeriodSeconds: & z.Spec.Pod.TerminationGracePeriodSeconds, 
	}
	if reflect.DeepEqual(v1.PodSecurityContext {}, z.Spec.Pod.SecurityContext) {
		podSpec.SecurityContext = z.Spec.Pod.SecurityContext
	}
	podSpec.NodeSelector = z.Spec.Pod.NodeSelector
	podSpec.Tolerations = z.Spec.Pod.Tolerations

	return podSpec
}

// MakeClientService returns a client service resource for the mosquitto cluster
func MakeClientService(z * v1alpha1.MosquittoCluster) * v1.Service {
	ports := z.MosquittoPorts()
	svcPorts := []v1.ServicePort { {Name:"client", Port:ports.Client}, 
	}
	return makeService(z.GetClientServiceName(), svcPorts, true, z)
}

// MakeConfigMap returns a mosquitto config map
func MakeConfigMap(z * v1alpha1.MosquittoCluster) * v1.ConfigMap {
	return & v1.ConfigMap {
		TypeMeta:metav1.TypeMeta {
			Kind:"ConfigMap", 
			APIVersion:"v1", 
		}, 
		ObjectMeta:metav1.ObjectMeta {
			Name:z.ConfigMapName(), 
			Namespace:z.Namespace, 
		}, 
		Data:map[string]string {
			"mosquitto.conf":makeMqConfigString(z), 
			"env.sh":makeMqEnvConfigString(z), 
		}, 
	}
}

// MakeHeadlessService returns an internal headless-service for the mosquitto statefulset
func MakeHeadlessService(z * v1alpha1.MosquittoCluster) * v1.Service {
	ports := z.MosquittoPorts()
	svcPorts := []v1.ServicePort { {Name:"websocket", Port:ports.Websocket},  {Name:"monitor", Port:ports.Monitor}, 
	}
	return makeService(headlessSvcName(z), svcPorts, false, z)
}

func makeMqConfigString(s *v1alpha1.MosquittoCluster)string {
	var result string
	for index := 1 ; index < int(s.Spec.Replicas) ; index++ {
		result +=  "connection myconn-" + strconv.Itoa(index)+ "\n\n"
                result +=  "address " + s.GetName() + "-" +  strconv.Itoa(index) + "." + headlessDomain(s) + "\n\n"
		result +=  "topic room1/# both 2 sensor/myhouse/\n\n"
		result +=  "bridge_protocol_version mqttv311\n\n"
		result +=  "notifications true\n\n"
		result +=  "cleansession true\n\n"
		result +=  "start_type automatic\n\n"
	}

	return result
}

func makeMqEnvConfigString(z * v1alpha1.MosquittoCluster)string {
	ports := z.MosquittoPorts()
	return "#!/usr/bin/env bash\n\n" + 
		"DOMAIN=" + headlessDomain(z) + "\n" + 
		"WEBSOCKET_PORT=" + strconv.Itoa(int(ports.Websocket)) + "\n" + 
		"MONITOR_PORT=" + strconv.Itoa(int(ports.Monitor)) + "\n" + 
		"CLIENT_HOST=" + z.GetClientServiceName() + "\n" + 
		"CLIENT_PORT=" + strconv.Itoa(int(ports.Client)) + "\n" + 
		"CLUSTER_SIZE=" + fmt.Sprint(z.Spec.Replicas) + "\n"
}

func makeService(name string, ports []v1.ServicePort, clusterIP bool, z * v1alpha1.MosquittoCluster) * v1.Service {
	service := v1.Service {
		TypeMeta:metav1.TypeMeta {
			Kind:"Service", 
			APIVersion:"v1", 
		}, 
		ObjectMeta:metav1.ObjectMeta {
			Name:name, 
			Namespace:z.Namespace, 
			OwnerReferences:[]metav1.OwnerReference { * metav1.NewControllerRef(z, schema.GroupVersionKind {
					Group:v1alpha1.SchemeGroupVersion.Group, 
					Version:v1alpha1.SchemeGroupVersion.Version, 
					Kind:"MosquittoCluster", 
				}), 
			}, 
			Labels:map[string]string {"app":z.GetName()}, 
		}, 
		Spec:v1.ServiceSpec {
			Ports:ports, 
			Selector:map[string]string {"app":z.GetName()}, 
		}, 
	}
	if clusterIP == false {
		service.Spec.ClusterIP = v1.ClusterIPNone
	}
	return & service
}

// MakePodDisruptionBudget returns a pdb for the mosquitto cluster
func MakePodDisruptionBudget(z * v1alpha1.MosquittoCluster) * policyv1beta1.PodDisruptionBudget {
	pdbCount := intstr.FromInt(1)
	return & policyv1beta1.PodDisruptionBudget {
		TypeMeta:metav1.TypeMeta {
			Kind:"PodDisruptionBudget", 
			APIVersion:"policy/v1beta1", 
		}, 
		ObjectMeta:metav1.ObjectMeta {
			Name:z.GetName(), 
			Namespace:z.Namespace, 
			OwnerReferences:[]metav1.OwnerReference { * metav1.NewControllerRef(z, schema.GroupVersionKind {
					Group:v1alpha1.SchemeGroupVersion.Group, 
					Version:v1alpha1.SchemeGroupVersion.Version, 
					Kind:"MosquittoCluster", 
				}), 
			}, 
		}, 
		Spec:policyv1beta1.PodDisruptionBudgetSpec {
			MaxUnavailable: & pdbCount, 
			Selector: & metav1.LabelSelector {
				MatchLabels:map[string]string {
					"app":z.GetName(), 
				}, 
			}, 
		}, 
	}
}
