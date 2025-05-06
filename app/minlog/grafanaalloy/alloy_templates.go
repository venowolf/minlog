package grafanaalloy

var alloy_template string = `
logging {
    level = "warn"
}

loki.write "{{ .NodeNameWithOutDash }}" {
    endpoint {
        url = "{{ .Lokiep }}"
    }
    external_labels = {
        node = "{{ .NodeName }}",
    }
}
`

var pod_template string = `
{{ range $k, $v := .Pod.ContainerMap }}
local.file_match "lf_{{ slice $v 0 16 }}" {
    path_targets = [{"__path__" = "{{ $.Pod.AppLogs }}/{{ $.Pod.PodName }}_{{ $.Pod.NameSpace }}_{{ $k }}-{{ $v }}.log"}]
}

loki.source.file "lsf_{{ slice $v 0 16 }}" {
    targets    = local.file_match.lf_{{ slice $v 0 16 }}.targets
    forward_to = [loki.process.lp_{{ slice $v 0 16 }}.receiver]
}

loki.process "lp_{{ slice $v 0 16 }}" {
    forward_to = [loki.write.{{ $.NodeNameWithOutDash }}.receiver]
    stage.static_labels {
        values = {
            service = "{{ $.Pod.ServiceName }}",
            podname = "{{ $.Pod.PodName }}",
            container = "{{ $k }}",
            namespace = "{{ $.Pod.NameSpace }}",
        }
    }
}
{{ end }}
`
