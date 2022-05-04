package consts

const (
	CreateKubernetesVerb = "create"
	UpdateKubernetesVerb = "update"
	PatchKubernetesVerb  = "patch"
	DeleteKubernetesVerb = "delete"
	GetKubernetesVerb    = "get"
	ListKubernetesVerb   = "list"
	WatchKubernetesVerb  = "watch"

	NamespacesKubernetesResource      = "namespaces"
	DeploymentsKubernetesResource     = "deployments"
	ServiceAccountsKubernetesResource = "serviceaccounts"
	RolesKubernetesResource           = "roles"
	RoleBindingsKubernetesResource    = "rolebindings"
	PodsKubernetesResource            = "pods"

	ClusterRoleKubernetesResourceType = "ClusterRole"

	AllApiGroups              = "*"
	RbacAuthorizationApiGroup = "rbac.authorization.k8s.io"
)
