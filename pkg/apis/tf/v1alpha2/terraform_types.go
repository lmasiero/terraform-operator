package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// Terraform is the Schema for the terraforms API
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=terraforms,shortName=tf
// +kubebuilder:singular=terraform
type Terraform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformSpec   `json:"spec,omitempty"`
	Status TerraformStatus `json:"status,omitempty"`
}

// TerraformSpec defines the desired state of Terraform
// +k8s:openapi-gen=true
type TerraformSpec struct {

	// KeepLatestPodsOnly when true will keep only the pods that match the
	// current generation of the terraform k8s-resource. This overrides the
	// behavior of `keepCompletedPods`.
	KeepLatestPodsOnly bool `json:"keepLatestPodsOnly,omitempty"`

	// KeepCompletedPods when true will keep completed pods. Default is false
	// and completed pods are removed.
	KeepCompletedPods bool `json:"keepCompletedPods,omitempty"`

	// OutputsSecret will create a secret with the outputs from the module. All
	// outputs from the module will be written to the secret unless the user
	// defines "outputsToInclude" or "outputsToOmit".
	OutputsSecret string `json:"outputsSecret,omitempty"`

	// OutputsToInclude is a whitelist of outputs to write when writing the
	// outputs to kubernetes.
	OutputsToInclude []string `json:"outputsToInclude,omitempty"`

	// OutputsToOmit is a blacklist of outputs to omit when writing the
	// outputs to kubernetes.
	OutputsToOmit []string `json:"outputsToOmit,omitempty"`

	// WriteOutputsToStatus will add the outputs from the module to the status
	// of the Terraform CustomResource.
	WriteOutputsToStatus bool `json:"writeOutputsToStatus,omitempty"`

	// PersistentVolumeSize define the size of the disk used to store
	// terraform run data. If not defined, a default of "2Gi" is used.
	PersistentVolumeSize *resource.Quantity `json:"persistentVolumeSize,omitempty"` // NOT MUTABLE

	// ServiceAccount use a specific kubernetes ServiceAccount for running the create + destroy pods.
	// If not specified we create a new ServiceAccount per Terraform
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Credentials is an array of credentials generally used for Terraform
	// providers
	Credentials []Credentials `json:"credentials,omitempty"`

	// IgnoreDelete will bypass the finalization process and remove the tf
	// resource without running any delete jobs.
	IgnoreDelete bool `json:"ignoreDelete,omitempty"`

	// SSHTunnel can be defined for pulling from scm sources that cannot be accessed by the network the
	// operator/runner runs in. An example is enterprise-Github servers running on a private network.
	SSHTunnel *ProxyOpts `json:"sshTunnel,omitempty"`

	// SCMAuthMethods define multiple SCMs that require tokens/keys
	SCMAuthMethods []SCMAuthMethod `json:"scmAuthMethods,omitempty"`

	// Images describes the container images used by task classes.
	Images *Images `json:"images,omitempty"`

	// Setup is configuration generally used once in the setup task
	Setup *Setup `json:"setup,omitempty"`

	// TerraformModule is used to configure the source of the terraform module.
	TerraformModule Module `json:"terraformModule"`

	// TerraformVersion is the version of terraform which is used to run the module. The terraform version is
	// used as the tag of the terraform image  regardless if images.terraform.image is defined with a tag. In
	// that case, the tag is stripped and replace with this value.
	TerraformVersion string `json:"terraformVersion"`

	// Backend is mandatory terraform backend configuration. Must use a valid terraform backend block.
	// For more information see https://www.terraform.io/language/settings/backends/configuration
	//
	// Example usage of the kubernetes cluster as a backend:
	//
	//   terraform {
	//    backend "kubernetes" {
	//     secret_suffix     = "all-task-types"
	//     namespace         = "default"
	//     in_cluster_config = true
	//    }
	//   }
	//
	// Example of a remote backend:
	//
	//   terraform {
	//    backend "remote" {
	//     organization = "example_corp"
	//     workspaces {
	//       name = "my-app-prod"
	//     }
	//    }
	//   }
	//
	//
	Backend string `json:"backend"`

	// TaskOptions are a list of configuration options to be injected into task pods.
	TaskOptions []TaskOption `json:"taskOptions,omitempty"`
}

// Setup are things that only happen during the life of the setup task.
type Setup struct {
	// ResourceDownloads defines other files to download into the module directory that can be used by the
	// terraform workflow runners. The `tfvar` type will also be fetched by the `exportRepo` option
	// (if defined) to aggregate the set of tfvars to save to an scm system.
	ResourceDownloads []ResourceDownload `json:"resourceDownloads,omitempty"`

	// CleanupDisk will clear out previous terraform run data from the persistent volume.
	CleanupDisk bool `json:"cleanupDisk,omitempty"`
}

// // Images describes the container images used by task classes
type Images struct {
	// Terraform task type container image definition
	Terraform *ImageConfig `json:"terraform,omitempty"`
	// Script task type container image definition
	Script *ImageConfig `json:"script,omitempty"`
	// Setup task type container image definition
	Setup *ImageConfig `json:"setup,omitempty"`
}

// ImageConfig describes a task class's container image and image pull policy.
type ImageConfig struct {

	// The container image from the registry; tags must be omitted
	Image string `json:"image"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
}

// Module has the different types of ways to define a terraform module. The order of precendence is
//     1. inline
//     2. configMapSelector
//     3. source[/version]
type Module struct {
	// Source accepts a subset of the terraform "Module Source" ways of defining a module.
	// Terraform Operator prefers modules that are defined in a git repo as opposed to other scm types.
	// Refer to https://www.terraform.io/language/modules/sources#module-sources for more details.
	Source string `json:"source,omitempty"`
	// Version to select from a terraform registry. For version to be used, source must be defined.
	// Refer to https://www.terraform.io/language/modules/sources#module-sources for more details
	Version string `json:"version,omitempty"`

	// ConfigMapSelector is an option that points to an existing configmap on the executing cluster. The
	// configmap is expected to contains has the terraform module (ie keys ending with .tf).
	// The configmap would need to live in the same namespace as the tfo resource.
	//
	// The configmap is mounted as a volume and put into the TFO_MAIN_MODULE path by the setup task.
	//
	// If a key is defined, the value is used as the module else the entirety of the data objects will be
	// loaded as files.
	ConfigMapSelector *ConfigMapSelector `json:"configMapSeclector,omitempty"`

	// Inline used to define an entire terraform module inline and then mounted in the TFO_MAIN_MODULE path.
	Inline string `json:"inline,omitempty"`
}

type TaskType string

func (t TaskType) String() string {
	return string(t)
}

const (
	RunSetupDelete     TaskType = "setup-delete"
	RunPreInitDelete   TaskType = "preinit-delete"
	RunInitDelete      TaskType = "init-delete"
	RunPostInitDelete  TaskType = "postinit-delete"
	RunPrePlanDelete   TaskType = "preplan-delete"
	RunPlanDelete      TaskType = "plan-delete"
	RunPostPlanDelete  TaskType = "postplan-delete"
	RunPreApplyDelete  TaskType = "preapply-delete"
	RunApplyDelete     TaskType = "apply-delete"
	RunPostApplyDelete TaskType = "postapply-delete"

	RunSetup     TaskType = "setup"
	RunPreInit   TaskType = "preinit"
	RunInit      TaskType = "init"
	RunPostInit  TaskType = "postinit"
	RunPrePlan   TaskType = "preplan"
	RunPlan      TaskType = "plan"
	RunPostPlan  TaskType = "postplan"
	RunPreApply  TaskType = "preapply"
	RunApply     TaskType = "apply"
	RunPostApply TaskType = "postapply"
	RunNil       TaskType = ""

	// RunExport RunType = "export"
)

// TaskOption are different configuration options to be injected into task pods. Can apply to
// one ore more task pods.
type TaskOption struct {
	// TaskTypes is a list of tasks these options will get applied to.
	TaskTypes []TaskType `json:"runTypes"`

	// RunnerRules are RBAC rules that will be added to all runner pods.
	PolicyRules []rbacv1.PolicyRule `json:"policyRules,omitempty"`

	// Labels extra labels to add task pods.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotaitons extra annotaitons to add the task pods
	Annotations map[string]string `json:"annotations,omitempty"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`

	// List of environment variables to set in the task pods.
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`

	// Compute Resources required by the task pods.
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Script is used to configure the source of the task's executable script.
	Script StageScript `json:"script,omitempty"`
}

// StageScript defines the different ways of sourcing execution scripts of tasks. There is an order of
// precendence of selecting which source is used, which is:
//     1. inline
//     2. configMapSelector
//     3. source
type StageScript struct {
	// Source is an http source that the task container will fetch and then execute.
	Source string `json:"source,omitempty"`

	// ConfigMapSelector reads a in a script from a configmap name+key
	ConfigMapSelector *ConfigMapSelector `json:"configMapSelector,omitempty"`

	// Inline is used to write the entire task execution script in the tfo resource.
	Inline string `json:"inline,omitempty"`
}

// A simple selector for configmaps that can select on the name of the configmap
// with the optional key. The namespace is not an option since only runners
// with a namespace'd role will utilize this map.
type ConfigMapSelector struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

// SCMAuthMethod definition of SCMs that require tokens/keys
type SCMAuthMethod struct {
	Host string `json:"host"`

	// Git configuration options for auth methods of git
	Git *GitSCM `json:"git,omitempty"`
}

// GitSCM define the auth methods of git
type GitSCM struct {
	SSH   *GitSSH   `json:"ssh,omitempty"`
	HTTPS *GitHTTPS `json:"https,omitempty"`
}

// GitSSH configurs the setup for git over ssh with optional proxy
type GitSSH struct {
	RequireProxy    bool             `json:"requireProxy,omitempty"`
	SSHKeySecretRef *SSHKeySecretRef `json:"sshKeySecretRef"`
}

// GitHTTPS configures the setup for git over https using tokens. Proxy is not
// supported in the terraform job pod at this moment
// TODO HTTPS Proxy support
type GitHTTPS struct {
	RequireProxy   bool            `json:"requireProxy,omitempty"`
	TokenSecretRef *TokenSecretRef `json:"tokenSecretRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TerraformList contains a list of Terraform
type TerraformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Terraform `json:"items"`
}

// ResourceDownload (formerly SrcOpts) defines a resource to fetch using one
// of the configured protocols: ssh|http|https (eg git::SSH or git::HTTPS)
type ResourceDownload struct {

	// Address defines the source address resources to fetch.
	Address string `json:"address"`

	// Path will download the resources into this path which is relative to
	// the main module directory.
	Path string `json:"path,omitempty"`

	// UseAsVar will add the file as a tfvar via the -var-file flag of the
	// terraform plan command. The downloaded resource must not be a directory.
	UseAsVar bool `json:"useAsVar,omitempty"`
}

// ProxyOpts configures ssh tunnel/socks5 for downloading ssh/https resources
type ProxyOpts struct {
	Host            string          `json:"host,omitempty"`
	User            string          `json:"user,omitempty"`
	SSHKeySecretRef SSHKeySecretRef `json:"sshKeySecretRef"`
}

// SSHKeySecretRef defines the secret where the SSH key (for the proxy, git, etc) is stored
type SSHKeySecretRef struct {
	// Name the secret name that has the SSH key
	Name string `json:"name"`
	// Namespace of the secret; Default is the namespace of the terraform resource
	Namespace string `json:"namespace,omitempty"`
	// Key in the secret ref. Default to `id_rsa`
	Key string `json:"key,omitempty"`
}

// TokenSecretRef defines the token or password that can be used to log into a system (eg git)
type TokenSecretRef struct {
	// Name the secret name that has the token or password
	Name string `json:"name"`
	// Namespace of the secret; Default is the namespace of the terraform resource
	Namespace string `json:"namespace,omitempty"`
	// Key in the secret ref. Default to `token`
	Key string `json:"key,omitempty"`
}

// Credentials are used for adding credentials for terraform providers.
// For example, in AWS, the AWS Terraform Provider uses the default credential chain
// of the AWS SDK, one of which are environment variables (eg AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY)
type Credentials struct {
	// SecretNameRef will load environment variables into the terraform runner
	// from a kubernetes secret
	SecretNameRef SecretNameRef `json:"secretNameRef,omitempty"`
	// AWSCredentials contains the different methods to load AWS credentials
	// for the Terraform AWS Provider. If using AWS_ACCESS_KEY_ID and/or environment
	// variables for credentials, use fromEnvs.
	AWSCredentials AWSCredentials `json:"aws,omitempty"`

	// ServiceAccountAnnotations allows the service account to be annotated with
	// cloud IAM roles such as Workload Identity on GCP
	ServiceAccountAnnotations map[string]string `json:"serviceAccountAnnotations,omitempty"`

	// TODO Add other commonly used cloud providers to this list
}

// AWSCredentials provides a few different k8s-specific methods of adding
// crednetials to pods. This includes KIAM and IRSA.
//
// To use environment variables, use a secretNameRef instead.
type AWSCredentials struct {
	// IRSA requires the irsa role-arn as the string input. This will create a
	// serice account named tf-<resource-name>. In order for the pod to be able to
	// use this role, the "Trusted Entity" of the IAM role must allow this
	// serice account name and namespace.
	//
	// Using a TrustEntity policy that includes "StringEquals" setting it as the serivce account name
	// is the most secure way to use IRSA.
	//
	// However, for a reusable policy consider "StringLike" with a few wildcards to make
	// the irsa role usable by pods created by terraform-operator. The example below is
	// pretty liberal, but will work for any pod created by the terraform-operator.
	//
	// {
	//   "Version": "2012-10-17",
	//   "Statement": [
	//     {
	//       "Effect": "Allow",
	//       "Principal": {
	//         "Federated": "${OIDC_ARN}"
	//       },
	//       "Action": "sts:AssumeRoleWithWebIdentity",
	//       "Condition": {
	//         "StringLike": {
	//           "${OIDC_URL}:sub": "system:serviceaccount:*:tf-*"
	//         }
	//       }
	//     }
	//   ]
	// }
	IRSA string `json:"irsa,omitempty"`

	// KIAM requires the kiam role-name as the string input. This will add the
	// correct annotation to the terraform execution pod
	KIAM string `json:"kiam,omitempty"`
}

// SecretNameRef is the name of the kubernetes secret to use
type SecretNameRef struct {
	// Name of the secret
	Name string `json:"name"`
	// Namespace of the secret; Defaults to namespace of the tf resource
	Namespace string `json:"namespace,omitempty"`
	// Key of the secret
	Key string `json:"key,omitempty"`
}

// TerraformStatus defines the observed state of Terraform
// +k8s:openapi-gen=true
type TerraformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// PodNamePrefix is used to identify this installation of the resource. For
	// very long resource names, like those greater than 220 characters, the
	// prefix ensures resource uniqueness for runners and other resources used
	// by the runner.
	// Another case for the pod name prefix is when rapidly deleteing a resource
	// and recreating it, the chance of recycling existing resources is reduced
	// to virtually nil.
	PodNamePrefix           string            `json:"podNamePrefix"`
	Phase                   StatusPhase       `json:"phase"`
	LastCompletedGeneration int64             `json:"lastCompletedGeneration"`
	Outputs                 map[string]string `json:"outputs,omitempty"`
	Stages                  []Stage           `json:"stages"`
	Stage                   Stage             `json:"stage"`

	// TODO maybe change this to
	// ExportReady bool - when try can run eport on it... no tracking on it
	// ExportStatus string - mostly the same thing, just easier to understand
	//
	// Or just move export to an entirely different controller. Accepts the same
	// fileds of export, and reads in tf resource as ref. Benifit will run in
	// foreround instead of background. The cons are a new controller to
	// maintain.

	// Status of export if used
	Exported Exported `json:"exported,omitempty"`
}

type Exported string

const (
	ExportedTrue       Exported = "true"
	ExportedFalse      Exported = "false"
	ExportedInProgress Exported = "in-progress"
	ExportedFailed     Exported = "failed"
	ExportedPending    Exported = "pending"
	ExportCreating     Exported = "creating"
)

type Stage struct {
	Generation int64      `json:"generation"`
	State      StageState `json:"state"`
	PodType    TaskType   `json:"podType"`

	// Interruptible is set to false when the pod should not be terminated
	// such as when doing a terraform apply
	Interruptible Interruptible `json:"interruptible"`
	Reason        string        `json:"reason"`
	StartTime     metav1.Time   `json:"startTime,omitempty"`
	StopTime      metav1.Time   `json:"stopTime,omitempty"`
}

type StatusPhase string

const (
	PhaseInitializing StatusPhase = "initializing"
	PhaseCompleted    StatusPhase = "completed"
	PhaseRunning      StatusPhase = "running"
	PhaseInitDelete   StatusPhase = "initializing-delete"
	PhaseDeleting     StatusPhase = "deleting"
	PhaseDeleted      StatusPhase = "deleted"
)

type StageState string

const (
	StateInitializing StageState = "initializing"
	StateComplete     StageState = "complete"
	StateFailed       StageState = "failed"
	StateInProgress   StageState = "in-progress"
	StateUnknown      StageState = "unknown"
)

type Interruptible bool

const (
	CanNotBeInterrupt Interruptible = false
	CanBeInterrupt    Interruptible = true
)

func init() {
	SchemeBuilder.Register(&Terraform{}, &TerraformList{})

}
