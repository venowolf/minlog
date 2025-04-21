package agentflow

var agentflow_template string = `
logging {
    level = "warn"
}

loki.write "{{ .Nodename }}" {
    endpoint {
        url = "{{ .Lokiep }}"
    }
    external_labels = {
        node = "{{ .Nodename }}",
    }
}
`

var pod_template string = `
local.file_match "{{ .Podname}}_{{ .Containername }}" {
    path_targets = [{"__path__" = "/var/log/containers/{{.Podname}}_{{.Namespace}}_{{.Containername}}-{{.Containerid}}.log"}]
}

loki.source.file " {{ .Podname}}_{{ .Containername }}" {
    targets    = local.file_match.{{ .Podname}}_{{ .Containername }}.targets
    forward_to = [loki.process.{{ .Codname}}_{{ .Containername }}.receiver]
}

loki.process "{{ .Podname}}_{{ .Containername }}" {
    forward_to = [loki.write.{{ .Lokiep }}.receiver]
    stage.static_labels {
        values = {
            service = "{{ .Serviceename }}",
            podname = "{{ .Podname }}",
            container = "{{ .Containername }}",
            namespace = "{{ .Namespace }}",
        }
    }
}
`
