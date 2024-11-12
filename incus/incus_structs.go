package incus

import "time"

type IncusInstances struct {
	Type       string   `json:"type"`
	Status     string   `json:"status"`
	StatusCode int      `json:"status_code"`
	Operation  string   `json:"operation"`
	ErrorCode  int      `json:"error_code"`
	Error      string   `json:"error"`
	Metadata   []string `json:"metadata"`
}

type IncusInstance struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Operation  string `json:"operation"`
	ErrorCode  int    `json:"error_code"`
	Error      string `json:"error"`
	Metadata   struct {
		Architecture string `json:"architecture"`
		Config       struct {
			ImageArchitecture           string `json:"image.architecture"`
			ImageDescription            string `json:"image.description"`
			ImageOs                     string `json:"image.os"`
			ImageRelease                string `json:"image.release"`
			ImageSerial                 string `json:"image.serial"`
			ImageType                   string `json:"image.type"`
			ImageVariant                string `json:"image.variant"`
			VolatileBaseImage           string `json:"volatile.base_image"`
			VolatileCloudInitInstanceID string `json:"volatile.cloud-init.instance-id"`
			VolatileEth0HostName        string `json:"volatile.eth0.host_name"`
			VolatileEth0Hwaddr          string `json:"volatile.eth0.hwaddr"`
			VolatileIdmapBase           string `json:"volatile.idmap.base"`
			VolatileIdmapCurrent        string `json:"volatile.idmap.current"`
			VolatileIdmapNext           string `json:"volatile.idmap.next"`
			VolatileLastStateIdmap      string `json:"volatile.last_state.idmap"`
			VolatileLastStatePower      string `json:"volatile.last_state.power"`
			VolatileUUID                string `json:"volatile.uuid"`
			VolatileUUIDGeneration      string `json:"volatile.uuid.generation"`
		} `json:"config"`
		Devices        struct{}  `json:"devices"`
		Ephemeral      bool      `json:"ephemeral"`
		Profiles       []string  `json:"profiles"`
		Stateful       bool      `json:"stateful"`
		Description    string    `json:"description"`
		CreatedAt      time.Time `json:"created_at"`
		ExpandedConfig struct {
			ImageArchitecture           string `json:"image.architecture"`
			ImageDescription            string `json:"image.description"`
			ImageOs                     string `json:"image.os"`
			ImageRelease                string `json:"image.release"`
			ImageSerial                 string `json:"image.serial"`
			ImageType                   string `json:"image.type"`
			ImageVariant                string `json:"image.variant"`
			VolatileBaseImage           string `json:"volatile.base_image"`
			VolatileCloudInitInstanceID string `json:"volatile.cloud-init.instance-id"`
			VolatileEth0HostName        string `json:"volatile.eth0.host_name"`
			VolatileEth0Hwaddr          string `json:"volatile.eth0.hwaddr"`
			VolatileIdmapBase           string `json:"volatile.idmap.base"`
			VolatileIdmapCurrent        string `json:"volatile.idmap.current"`
			VolatileIdmapNext           string `json:"volatile.idmap.next"`
			VolatileLastStateIdmap      string `json:"volatile.last_state.idmap"`
			VolatileLastStatePower      string `json:"volatile.last_state.power"`
			VolatileUUID                string `json:"volatile.uuid"`
			VolatileUUIDGeneration      string `json:"volatile.uuid.generation"`
		} `json:"expanded_config"`
		ExpandedDevices struct {
			Eth0 struct {
				Name    string `json:"name"`
				Network string `json:"network"`
				Type    string `json:"type"`
			} `json:"eth0"`
			Root struct {
				Path string `json:"path"`
				Pool string `json:"pool"`
				Type string `json:"type"`
			} `json:"root"`
		} `json:"expanded_devices"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		StatusCode int       `json:"status_code"`
		LastUsedAt time.Time `json:"last_used_at"`
		Location   string    `json:"location"`
		Type       string    `json:"type"`
		Project    string    `json:"project"`
	} `json:"metadata"`
}

type IncusInstanceStatus struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Operation  string `json:"operation"`
	ErrorCode  int    `json:"error_code"`
	Error      string `json:"error"`
	Metadata   struct {
		Status     string `json:"status"`
		StatusCode int    `json:"status_code"`
		Disk       struct {
			Root struct {
				Usage int `json:"usage"`
				Total int `json:"total"`
			} `json:"root"`
		} `json:"disk"`
		Memory struct {
			Usage         int   `json:"usage"`
			UsagePeak     int   `json:"usage_peak"`
			Total         int64 `json:"total"`
			SwapUsage     int   `json:"swap_usage"`
			SwapUsagePeak int   `json:"swap_usage_peak"`
		} `json:"memory"`
		Network struct {
			Eth0 struct {
				Addresses []struct {
					Family  string `json:"family"`
					Address string `json:"address"`
					Netmask string `json:"netmask"`
					Scope   string `json:"scope"`
				} `json:"addresses"`
				Counters struct {
					BytesReceived          int `json:"bytes_received"`
					BytesSent              int `json:"bytes_sent"`
					PacketsReceived        int `json:"packets_received"`
					PacketsSent            int `json:"packets_sent"`
					ErrorsReceived         int `json:"errors_received"`
					ErrorsSent             int `json:"errors_sent"`
					PacketsDroppedOutbound int `json:"packets_dropped_outbound"`
					PacketsDroppedInbound  int `json:"packets_dropped_inbound"`
				} `json:"counters"`
				Hwaddr   string `json:"hwaddr"`
				HostName string `json:"host_name"`
				Mtu      int    `json:"mtu"`
				State    string `json:"state"`
				Type     string `json:"type"`
			} `json:"eth0"`
			Lo struct {
				Addresses []struct {
					Family  string `json:"family"`
					Address string `json:"address"`
					Netmask string `json:"netmask"`
					Scope   string `json:"scope"`
				} `json:"addresses"`
				Counters struct {
					BytesReceived          int `json:"bytes_received"`
					BytesSent              int `json:"bytes_sent"`
					PacketsReceived        int `json:"packets_received"`
					PacketsSent            int `json:"packets_sent"`
					ErrorsReceived         int `json:"errors_received"`
					ErrorsSent             int `json:"errors_sent"`
					PacketsDroppedOutbound int `json:"packets_dropped_outbound"`
					PacketsDroppedInbound  int `json:"packets_dropped_inbound"`
				} `json:"counters"`
				Hwaddr   string `json:"hwaddr"`
				HostName string `json:"host_name"`
				Mtu      int    `json:"mtu"`
				State    string `json:"state"`
				Type     string `json:"type"`
			} `json:"lo"`
		} `json:"network"`
		Pid       int `json:"pid"`
		Processes int `json:"processes"`
		CPU       struct {
			Usage int64 `json:"usage"`
		} `json:"cpu"`
	} `json:"metadata"`
}

type IncusStateResponse struct {
	Type       string `json:"type,omitempty"`
	Status     string `json:"status,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Operation  string `json:"operation,omitempty"`
	ErrorCode  int    `json:"error_code,omitempty"`
	Error      string `json:"error,omitempty"`
	Metadata   struct {
		ID          string `json:"id,omitempty"`
		Class       string `json:"class,omitempty"`
		Description string `json:"description,omitempty"`
		CreatedAt   string `json:"created_at,omitempty"`
		UpdatedAt   string `json:"updated_at,omitempty"`
		Status      string `json:"status,omitempty"`
		StatusCode  int    `json:"status_code,omitempty"`
		Resources   struct {
			Containers []string `json:"containers,omitempty"`
			Instances  []string `json:"instances,omitempty"`
		} `json:"resources,omitempty"`
		Metadata  any    `json:"metadata,omitempty"`
		MayCancel bool   `json:"may_cancel,omitempty"`
		Err       string `json:"err,omitempty"`
		Location  string `json:"location,omitempty"`
	} `json:"metadata,omitempty"`
}

type IncusCreate struct {
	Architecture string `json:"architecture,omitempty"`
	Config       struct {
		SecurityNesting string `json:"security.nesting,omitempty"`
	} `json:"config,omitempty"`
	Description string `json:"description,omitempty"`
	Devices     struct {
		Root struct {
			Path string `json:"path,omitempty"`
			Pool string `json:"pool,omitempty"`
			Type string `json:"type,omitempty"`
		} `json:"root,omitempty"`
	} `json:"devices,omitempty"`
	Ephemeral    bool     `json:"ephemeral,omitempty"`
	InstanceType string   `json:"instance_type,omitempty"`
	Name         string   `json:"name,omitempty"`
	Profiles     []string `json:"profiles,omitempty"`
	Restore      string   `json:"restore,omitempty"`
	Source       struct {
		Alias             string `json:"alias,omitempty"`
		AllowInconsistent bool   `json:"allow_inconsistent,omitempty"`
		BaseImage         string `json:"base-image,omitempty"`
		Certificate       string `json:"certificate,omitempty"`
		Fingerprint       string `json:"fingerprint,omitempty"`
		InstanceOnly      bool   `json:"instance_only,omitempty"`
		Live              bool   `json:"live,omitempty"`
		Mode              string `json:"mode,omitempty"`
		Operation         string `json:"operation,omitempty"`
		Project           string `json:"project,omitempty"`
		Properties        struct {
			Os      string `json:"os,omitempty"`
			Release string `json:"release,omitempty"`
			Variant string `json:"variant,omitempty"`
		} `json:"properties,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		Refresh  bool   `json:"refresh,omitempty"`
		Secret   string `json:"secret,omitempty"`
		Secrets  struct {
			Criu  string `json:"criu,omitempty"`
			Rsync string `json:"rsync,omitempty"`
		} `json:"secrets,omitempty"`
		Server string `json:"server,omitempty"`
		Source string `json:"source,omitempty"`
		Type   string `json:"type,omitempty"`
	} `json:"source,omitempty"`
	Start    bool   `json:"start,omitempty"`
	Stateful bool   `json:"stateful,omitempty"`
	Type     string `json:"type,omitempty"`
}

type IncusCopy struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Operation  string `json:"operation"`
	ErrorCode  int    `json:"error_code"`
	Error      string `json:"error"`
	Metadata   struct {
		ID          string `json:"id"`
		Class       string `json:"class"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Status      string `json:"status"`
		StatusCode  int    `json:"status_code"`
		Resources   struct {
			Containers []string `json:"containers"`
			Instances  []string `json:"instances"`
		} `json:"resources"`
		Metadata  any    `json:"metadata"`
		MayCancel bool   `json:"may_cancel"`
		Err       string `json:"err"`
		Location  string `json:"location"`
	} `json:"metadata"`
}

type IncusConsole struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Operation  string `json:"operation"`
	ErrorCode  int    `json:"error_code"`
	Error      string `json:"error"`
	Metadata   struct {
		ID          string `json:"id"`
		Class       string `json:"class"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Status      string `json:"status"`
		StatusCode  int    `json:"status_code"`
		Resources   struct {
			Containers []string `json:"containers"`
			Instances  []string `json:"instances"`
		} `json:"resources"`
		Metadata struct {
			Fds struct {
				Num0    string `json:"0"`
				Control string `json:"control"`
			} `json:"fds"`
		} `json:"metadata"`
		MayCancel bool   `json:"may_cancel"`
		Err       string `json:"err"`
		Location  string `json:"location"`
	} `json:"metadata"`
}
