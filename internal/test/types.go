package test

// Response types for mock Podman/Docker API server.
// These types are designed to be compatible with both Libpod and Docker APIs.

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Cause    string `json:"cause,omitempty"`
	Message  string `json:"message"`
	Response int    `json:"response"`
}

// VersionResponse represents the response from /version endpoint.
type VersionResponse struct {
	APIVersion    string             `json:"ApiVersion"`
	Arch          string             `json:"Arch"`
	Built         int64              `json:"Built,omitempty"`
	BuildTime     string             `json:"BuildTime,omitempty"`
	Components    []ComponentVersion `json:"Components,omitempty"`
	Experimental  bool               `json:"Experimental,omitempty"`
	GitCommit     string             `json:"GitCommit"`
	GoVersion     string             `json:"GoVersion"`
	KernelVersion string             `json:"KernelVersion,omitempty"`
	MinAPIVersion string             `json:"MinAPIVersion,omitempty"`
	Os            string             `json:"Os"`
	Version       string             `json:"Version"`
}

// ComponentVersion represents a component in the version response.
type ComponentVersion struct {
	Details map[string]string `json:"Details,omitempty"`
	Name    string            `json:"Name"`
	Version string            `json:"Version"`
}

// InfoResponse represents the response from /info endpoint.
type InfoResponse struct {
	Host    HostInfo    `json:"host"`
	Store   StoreInfo   `json:"store"`
	Version VersionInfo `json:"version"`
}

// HostInfo contains host system information.
type HostInfo struct {
	Arch           string           `json:"arch"`
	BuildahVersion string           `json:"buildahVersion"`
	Conmon         ConmonInfo       `json:"conmon"`
	Distribution   DistributionInfo `json:"distribution"`
	Hostname       string           `json:"hostname"`
	Kernel         string           `json:"kernel"`
	MemFree        int64            `json:"memFree"`
	MemTotal       int64            `json:"memTotal"`
	OS             string           `json:"os"`
	Rootless       bool             `json:"rootless"`
	Uptime         string           `json:"uptime"`
	CPUs           int              `json:"cpus"`
	EventLogger    string           `json:"eventLogger"`
	SecurityInfo   SecurityInfo     `json:"security"`
}

// ConmonInfo contains conmon information.
type ConmonInfo struct {
	Package string `json:"package"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

// DistributionInfo contains OS distribution information.
type DistributionInfo struct {
	Distribution string `json:"distribution"`
	Version      string `json:"version"`
}

// SecurityInfo contains security-related information.
type SecurityInfo struct {
	AppArmorEnabled    bool   `json:"apparmorEnabled"`
	SELinuxEnabled     bool   `json:"selinuxEnabled"`
	SeccompEnabled     bool   `json:"seccompEnabled"`
	SeccompProfilePath string `json:"seccompProfilePath"`
	Rootless           bool   `json:"rootless"`
}

// StoreInfo contains storage information.
type StoreInfo struct {
	GraphDriverName string             `json:"graphDriverName"`
	GraphRoot       string             `json:"graphRoot"`
	RunRoot         string             `json:"runRoot"`
	ImageStore      ImageStoreInfo     `json:"imageStore"`
	ContainerStore  ContainerStoreInfo `json:"containerStore"`
}

// ImageStoreInfo contains image store statistics.
type ImageStoreInfo struct {
	Number int `json:"number"`
}

// ContainerStoreInfo contains container store statistics.
type ContainerStoreInfo struct {
	Number  int `json:"number"`
	Running int `json:"running"`
	Paused  int `json:"paused"`
	Stopped int `json:"stopped"`
}

// VersionInfo contains version information for the info endpoint.
type VersionInfo struct {
	APIVersion string `json:"APIVersion"`
	Built      int64  `json:"Built"`
	BuiltTime  string `json:"BuiltTime"`
	GitCommit  string `json:"GitCommit"`
	GoVersion  string `json:"GoVersion"`
	OsArch     string `json:"OsArch"`
	Version    string `json:"Version"`
}

// ContainerListResponse represents a container in the list response.
// Compatible with both Libpod and Docker APIs.
type ContainerListResponse struct {
	AutoRemove bool              `json:"AutoRemove,omitempty"`
	Command    []string          `json:"Command,omitempty"`
	Created    int64             `json:"Created"`
	CreatedAt  string            `json:"CreatedAt,omitempty"`
	ExitCode   int               `json:"ExitCode,omitempty"`
	Exited     bool              `json:"Exited,omitempty"`
	ExitedAt   int64             `json:"ExitedAt,omitempty"`
	ID         string            `json:"Id"`
	Image      string            `json:"Image"`
	ImageID    string            `json:"ImageID,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty"`
	Mounts     []string          `json:"Mounts,omitempty"`
	Names      []string          `json:"Names"`
	Namespaces struct{}          `json:"Namespaces,omitempty"`
	Networks   []string          `json:"Networks,omitempty"`
	Pid        int               `json:"Pid,omitempty"`
	Pod        string            `json:"Pod,omitempty"`
	PodName    string            `json:"PodName,omitempty"`
	Ports      []PortMapping     `json:"Ports,omitempty"`
	Size       *ContainerSize    `json:"Size,omitempty"`
	StartedAt  int64             `json:"StartedAt,omitempty"`
	State      string            `json:"State"`
	Status     string            `json:"Status,omitempty"`
}

// PortMapping represents a port mapping for a container.
type PortMapping struct {
	ContainerPort int    `json:"container_port,omitempty"`
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      int    `json:"host_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	// Docker-compatible fields
	IP          string `json:"IP,omitempty"`
	PrivatePort int    `json:"PrivatePort,omitempty"`
	PublicPort  int    `json:"PublicPort,omitempty"`
	Type        string `json:"Type,omitempty"`
}

// ContainerSize represents the size of a container.
type ContainerSize struct {
	RootFsSize int64 `json:"rootFsSize"`
	RwSize     int64 `json:"rwSize"`
}

// ContainerInspectResponse represents the detailed container information.
// Compatible with both Libpod and Docker APIs.
type ContainerInspectResponse struct {
	AppArmorProfile string           `json:"AppArmorProfile,omitempty"`
	Args            []string         `json:"Args,omitempty"`
	Config          *ContainerConfig `json:"Config,omitempty"`
	Created         string           `json:"Created"`
	Driver          string           `json:"Driver,omitempty"`
	HostConfig      *HostConfig      `json:"HostConfig,omitempty"`
	HostnamePath    string           `json:"HostnamePath,omitempty"`
	HostsPath       string           `json:"HostsPath,omitempty"`
	ID              string           `json:"Id"`
	Image           string           `json:"Image"`
	ImageName       string           `json:"ImageName,omitempty"`
	LogPath         string           `json:"LogPath,omitempty"`
	MountLabel      string           `json:"MountLabel,omitempty"`
	Mounts          []MountPoint     `json:"Mounts,omitempty"`
	Name            string           `json:"Name"`
	Namespace       string           `json:"Namespace,omitempty"`
	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty"`
	Path            string           `json:"Path,omitempty"`
	Platform        string           `json:"Platform,omitempty"`
	ProcessLabel    string           `json:"ProcessLabel,omitempty"`
	ResolvConfPath  string           `json:"ResolvConfPath,omitempty"`
	RestartCount    int              `json:"RestartCount,omitempty"`
	Rootfs          string           `json:"Rootfs,omitempty"`
	State           *ContainerState  `json:"State,omitempty"`
	StaticDir       string           `json:"StaticDir,omitempty"`
}

// ContainerConfig represents container configuration.
type ContainerConfig struct {
	AttachStderr bool                `json:"AttachStderr,omitempty"`
	AttachStdin  bool                `json:"AttachStdin,omitempty"`
	AttachStdout bool                `json:"AttachStdout,omitempty"`
	Cmd          []string            `json:"Cmd,omitempty"`
	Domainname   string              `json:"Domainname,omitempty"`
	Entrypoint   []string            `json:"Entrypoint,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	Hostname     string              `json:"Hostname,omitempty"`
	Image        string              `json:"Image,omitempty"`
	Labels       map[string]string   `json:"Labels,omitempty"`
	OnBuild      []string            `json:"OnBuild,omitempty"`
	OpenStdin    bool                `json:"OpenStdin,omitempty"`
	StdinOnce    bool                `json:"StdinOnce,omitempty"`
	StopSignal   string              `json:"StopSignal,omitempty"`
	StopTimeout  *int                `json:"StopTimeout,omitempty"`
	Tty          bool                `json:"Tty,omitempty"`
	User         string              `json:"User,omitempty"`
	Volumes      map[string]struct{} `json:"Volumes,omitempty"`
	WorkingDir   string              `json:"WorkingDir,omitempty"`
}

// HostConfig represents container host configuration.
type HostConfig struct {
	AutoRemove      bool                     `json:"AutoRemove,omitempty"`
	Binds           []string                 `json:"Binds,omitempty"`
	CapAdd          []string                 `json:"CapAdd,omitempty"`
	CapDrop         []string                 `json:"CapDrop,omitempty"`
	DNS             []string                 `json:"Dns,omitempty"`
	DNSOptions      []string                 `json:"DnsOptions,omitempty"`
	DNSSearch       []string                 `json:"DnsSearch,omitempty"`
	ExtraHosts      []string                 `json:"ExtraHosts,omitempty"`
	NetworkMode     string                   `json:"NetworkMode,omitempty"`
	PortBindings    map[string][]PortBinding `json:"PortBindings,omitempty"`
	Privileged      bool                     `json:"Privileged,omitempty"`
	PublishAllPorts bool                     `json:"PublishAllPorts,omitempty"`
	ReadonlyRootfs  bool                     `json:"ReadonlyRootfs,omitempty"`
	RestartPolicy   *RestartPolicy           `json:"RestartPolicy,omitempty"`
	SecurityOpt     []string                 `json:"SecurityOpt,omitempty"`
}

// PortBinding represents a port binding configuration.
type PortBinding struct {
	HostIP   string `json:"HostIp,omitempty"`
	HostPort string `json:"HostPort,omitempty"`
}

// RestartPolicy represents a container restart policy.
type RestartPolicy struct {
	MaximumRetryCount int    `json:"MaximumRetryCount,omitempty"`
	Name              string `json:"Name,omitempty"`
}

// MountPoint represents a mount point.
type MountPoint struct {
	Destination string `json:"Destination,omitempty"`
	Driver      string `json:"Driver,omitempty"`
	Mode        string `json:"Mode,omitempty"`
	Name        string `json:"Name,omitempty"`
	Propagation string `json:"Propagation,omitempty"`
	RW          bool   `json:"RW,omitempty"`
	Source      string `json:"Source,omitempty"`
	Type        string `json:"Type,omitempty"`
}

// NetworkSettings represents container network settings.
type NetworkSettings struct {
	Bridge                 string                    `json:"Bridge,omitempty"`
	EndpointID             string                    `json:"EndpointID,omitempty"`
	Gateway                string                    `json:"Gateway,omitempty"`
	GlobalIPv6Address      string                    `json:"GlobalIPv6Address,omitempty"`
	GlobalIPv6PrefixLen    int                       `json:"GlobalIPv6PrefixLen,omitempty"`
	HairpinMode            bool                      `json:"HairpinMode,omitempty"`
	IPAddress              string                    `json:"IPAddress,omitempty"`
	IPPrefixLen            int                       `json:"IPPrefixLen,omitempty"`
	IPv6Gateway            string                    `json:"IPv6Gateway,omitempty"`
	LinkLocalIPv6Address   string                    `json:"LinkLocalIPv6Address,omitempty"`
	LinkLocalIPv6PrefixLen int                       `json:"LinkLocalIPv6PrefixLen,omitempty"`
	MacAddress             string                    `json:"MacAddress,omitempty"`
	Networks               map[string]*NetworkDetail `json:"Networks,omitempty"`
	Ports                  map[string][]PortBinding  `json:"Ports,omitempty"`
	SandboxID              string                    `json:"SandboxID,omitempty"`
	SandboxKey             string                    `json:"SandboxKey,omitempty"`
}

// NetworkDetail represents network details for a container.
type NetworkDetail struct {
	Aliases             []string          `json:"Aliases,omitempty"`
	DriverOpts          map[string]string `json:"DriverOpts,omitempty"`
	EndpointID          string            `json:"EndpointID,omitempty"`
	Gateway             string            `json:"Gateway,omitempty"`
	GlobalIPv6Address   string            `json:"GlobalIPv6Address,omitempty"`
	GlobalIPv6PrefixLen int               `json:"GlobalIPv6PrefixLen,omitempty"`
	IPAMConfig          *IPAMConfig       `json:"IPAMConfig,omitempty"`
	IPAddress           string            `json:"IPAddress,omitempty"`
	IPPrefixLen         int               `json:"IPPrefixLen,omitempty"`
	IPv6Gateway         string            `json:"IPv6Gateway,omitempty"`
	Links               []string          `json:"Links,omitempty"`
	MacAddress          string            `json:"MacAddress,omitempty"`
	NetworkID           string            `json:"NetworkID,omitempty"`
}

// IPAMConfig represents IPAM configuration.
type IPAMConfig struct {
	IPv4Address  string   `json:"IPv4Address,omitempty"`
	IPv6Address  string   `json:"IPv6Address,omitempty"`
	LinkLocalIPs []string `json:"LinkLocalIPs,omitempty"`
}

// ContainerState represents the state of a container.
type ContainerState struct {
	Dead       bool         `json:"Dead,omitempty"`
	Error      string       `json:"Error,omitempty"`
	ExitCode   int          `json:"ExitCode,omitempty"`
	FinishedAt string       `json:"FinishedAt,omitempty"`
	Health     *HealthState `json:"Health,omitempty"`
	OOMKilled  bool         `json:"OOMKilled,omitempty"`
	Paused     bool         `json:"Paused,omitempty"`
	Pid        int          `json:"Pid,omitempty"`
	Restarting bool         `json:"Restarting,omitempty"`
	Running    bool         `json:"Running,omitempty"`
	StartedAt  string       `json:"StartedAt,omitempty"`
	Status     string       `json:"Status,omitempty"`
}

// HealthState represents the health state of a container.
type HealthState struct {
	FailingStreak int         `json:"FailingStreak,omitempty"`
	Log           []HealthLog `json:"Log,omitempty"`
	Status        string      `json:"Status,omitempty"`
}

// HealthLog represents a health check log entry.
type HealthLog struct {
	End      string `json:"End,omitempty"`
	ExitCode int    `json:"ExitCode,omitempty"`
	Output   string `json:"Output,omitempty"`
	Start    string `json:"Start,omitempty"`
}

// ImageListResponse represents an image in the list response.
// Compatible with both Libpod and Docker APIs.
type ImageListResponse struct {
	Containers  int               `json:"Containers,omitempty"`
	Created     int64             `json:"Created"`
	Dangling    bool              `json:"Dangling,omitempty"`
	Digest      string            `json:"Digest,omitempty"`
	History     []string          `json:"History,omitempty"`
	ID          string            `json:"Id"`
	Labels      map[string]string `json:"Labels,omitempty"`
	Names       []string          `json:"Names,omitempty"`
	ParentId    string            `json:"ParentId,omitempty"`
	ReadOnly    bool              `json:"ReadOnly,omitempty"`
	RepoDigests []string          `json:"RepoDigests,omitempty"`
	RepoTags    []string          `json:"RepoTags,omitempty"`
	SharedSize  int64             `json:"SharedSize,omitempty"`
	Size        int64             `json:"Size"`
	VirtualSize int64             `json:"VirtualSize,omitempty"`
}

// NetworkListResponse represents a network in the list response.
// Compatible with both Libpod and Docker APIs.
type NetworkListResponse struct {
	Attachable bool                   `json:"Attachable,omitempty"`
	ConfigFrom *ConfigReference       `json:"ConfigFrom,omitempty"`
	ConfigOnly bool                   `json:"ConfigOnly,omitempty"`
	Containers map[string]interface{} `json:"Containers,omitempty"`
	Created    string                 `json:"Created,omitempty"`
	Driver     string                 `json:"Driver,omitempty"`
	EnableIPv6 bool                   `json:"EnableIPv6,omitempty"`
	ID         string                 `json:"Id,omitempty"`
	IPAM       *IPAM                  `json:"IPAM,omitempty"`
	Ingress    bool                   `json:"Ingress,omitempty"`
	Internal   bool                   `json:"Internal,omitempty"`
	Labels     map[string]string      `json:"Labels,omitempty"`
	Name       string                 `json:"Name"`
	Options    map[string]string      `json:"Options,omitempty"`
	Scope      string                 `json:"Scope,omitempty"`
	// Libpod-specific fields
	DNSEnabled       bool     `json:"dns_enabled,omitempty"`
	IPV6Enabled      bool     `json:"ipv6_enabled,omitempty"`
	NetworkInterface string   `json:"network_interface,omitempty"`
	Subnets          []Subnet `json:"subnets,omitempty"`
}

// ConfigReference represents a network config reference.
type ConfigReference struct {
	Network string `json:"Network,omitempty"`
}

// IPAM represents IP Address Management configuration.
type IPAM struct {
	Config  []IPAMPoolConfig  `json:"Config,omitempty"`
	Driver  string            `json:"Driver,omitempty"`
	Options map[string]string `json:"Options,omitempty"`
}

// IPAMPoolConfig represents an IPAM pool configuration.
type IPAMPoolConfig struct {
	AuxiliaryAddresses map[string]string `json:"AuxiliaryAddresses,omitempty"`
	Gateway            string            `json:"Gateway,omitempty"`
	IPRange            string            `json:"IPRange,omitempty"`
	Subnet             string            `json:"Subnet,omitempty"`
}

// Subnet represents a network subnet.
type Subnet struct {
	Gateway string `json:"gateway,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
}

// VolumeListResponse represents the volume list response.
type VolumeListResponse struct {
	Volumes  []VolumeResponse `json:"Volumes,omitempty"`
	Warnings []string         `json:"Warnings,omitempty"`
}

// VolumeResponse represents a volume in the list response.
// Compatible with both Libpod and Docker APIs.
type VolumeResponse struct {
	CreatedAt  string                 `json:"CreatedAt,omitempty"`
	Driver     string                 `json:"Driver"`
	Labels     map[string]string      `json:"Labels,omitempty"`
	Mountpoint string                 `json:"Mountpoint"`
	Name       string                 `json:"Name"`
	Options    map[string]string      `json:"Options,omitempty"`
	Scope      string                 `json:"Scope,omitempty"`
	Status     map[string]interface{} `json:"Status,omitempty"`
	UsageData  *VolumeUsageData       `json:"UsageData,omitempty"`
	// Libpod-specific fields
	Anonymous   bool `json:"Anonymous,omitempty"`
	GID         int  `json:"GID,omitempty"`
	UID         int  `json:"UID,omitempty"`
	NeedsChown  bool `json:"NeedsChown,omitempty"`
	NeedsCopyUp bool `json:"NeedsCopyUp,omitempty"`
}

// VolumeUsageData represents volume usage data.
type VolumeUsageData struct {
	RefCount int64 `json:"RefCount,omitempty"`
	Size     int64 `json:"Size,omitempty"`
}

// ContainerCreateRequest represents a container creation request.
type ContainerCreateRequest struct {
	Name       string            `json:"name,omitempty"`
	Image      string            `json:"image"`
	Cmd        []string          `json:"command,omitempty"`
	Env        []string          `json:"env,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	HostConfig *HostConfig       `json:"HostConfig,omitempty"`
}

// ContainerCreateResponse represents a container creation response.
type ContainerCreateResponse struct {
	ID       string   `json:"Id"`
	Warnings []string `json:"Warnings,omitempty"`
}

// ImagePullResponse represents an image pull progress response.
type ImagePullResponse struct {
	Error          string          `json:"error,omitempty"`
	ErrorDetail    *ErrorDetail    `json:"errorDetail,omitempty"`
	ID             string          `json:"id,omitempty"`
	Images         []string        `json:"images,omitempty"`
	Progress       string          `json:"progress,omitempty"`
	ProgressDetail *ProgressDetail `json:"progressDetail,omitempty"`
	Status         string          `json:"status,omitempty"`
	Stream         string          `json:"stream,omitempty"`
}

// ErrorDetail represents error details in a response.
type ErrorDetail struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ProgressDetail represents progress details.
type ProgressDetail struct {
	Current int64 `json:"current,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

// ImageBuildResponse represents an image build progress response.
type ImageBuildResponse struct {
	Aux         *BuildAux    `json:"aux,omitempty"`
	Error       string       `json:"error,omitempty"`
	ErrorDetail *ErrorDetail `json:"errorDetail,omitempty"`
	ID          string       `json:"id,omitempty"`
	Progress    string       `json:"progress,omitempty"`
	Status      string       `json:"status,omitempty"`
	Stream      string       `json:"stream,omitempty"`
}

// BuildAux represents auxiliary build information.
type BuildAux struct {
	ID string `json:"ID,omitempty"`
}

// ImageRemoveResponse represents the response from removing an image.
type ImageRemoveResponse struct {
	Deleted  string `json:"Deleted,omitempty"`
	Untagged string `json:"Untagged,omitempty"`
}
