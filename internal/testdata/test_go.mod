module github.com/checkmarx/ast-cli

        go 1.24.2

        require (
        github.com/Checkmarx/containers-resolver v1.0.9
        github.com/Checkmarx/gen-ai-prompts v0.0.0-20240807143411-708ceec12b63
        gotest.tools v2.2.0+incompatible
        )

        require (
        dario.cat/mergo v1.0.1 // indirect
        k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
        sigs.k8s.io/yaml v1.4.0 // indirect
        )

        replace github.com/containerd/containerd => github.com/containerd/containerd v1.7.27
