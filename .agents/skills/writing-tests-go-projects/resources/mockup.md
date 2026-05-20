# mockup

Using Mockery to generate mocks for testing in Go projects. Use template `testify` for compatibility with the testify assertion library.


## Instructions
1. Install mockery: `go install github.com/vektra/mockery/v3@v3.7.0`
2. Initialize mockery in your project: `mockery init [module_name]`
3. Update the generated `mockery.yaml` configuration file to specify which interfaces to mock and where to output the generated files.
4. Run `mockery generate` to generate the mock files based on your configuration.

- All mocks will be generated in the `mocks` directory have folder structure same as the original interfaces.
- The generated mock files will be named `mocks_[interface_name]_test.go` 
- The generated mock struct names will be in the format `Mock[InterfaceName]` (e.g. `MockFoo` for interface `Foo`).


## Configuration

mockery init [module_name] is a command can bootstrap you with a fully-functioning configuration set.

```bash
mockery init myproject/mypackage
```

A more complex configuration example can be seen below:
```yaml
    all: False
    template-data:
      boilerplate-file: ./path/to/boilerplate.txt
    template: testify
    packages:
      github.com/vektra/example:
        config:
          # Make use of the template variables to place the mock in the same
          # directory as the original interface.
          dir: "{{.InterfaceDir}}"
          filename: "mocks_test.go"
          pkgname: "{{.PackageName}}_test"
          structname: "{{.Mock}}{{.InterfaceName}}"
        interfaces:
          Foo:
          Bar:
            config:
              # Make it unexported instead
              structname: "mock{{.InterfaceName}}"
          Baz:
            # Create two mock implementations of Baz with different names.
            configs:
              - filename: "mocks_baz_one_test.go"
                structname: "MockBazOne"
              - filename: "mocks_baz_two_test.go"
                structname: "MockBazTwo"
      io:
        config:
          dir: path/to/io/mocks
          filename: "mocks_io.go"
```
Parameter Descriptions
-----------------------
| name                                                   | templated                 | default                               | description                                                                                                                                                                                                                                          |
|--------------------------------------------------------|---------------------------|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `all`                                                  | :fontawesome-solid-x:     | `false`                        | Generate all interfaces for the specified packages.                                                                                                                                                                                                  |
| `_anchors`                                             | :fontawesome-solid-x:     | `{}`                           | Unused by mockery, but allowed in the config schema so that you may define arbitrary yaml anchors.                                                                                                                                                   |
| `config`                                               | :fontawesome-solid-x:     | `""`                           | Set the location of the mockery config file.                                                                                                                                                                                                         |
| `dir`                                                  | :fontawesome-solid-check: | `"{{.InterfaceDir}}"`  | The directory where the mock file will be outputted to.                                                                                                                                                                                              |
| `exclude-subpkg-regex`                                 | :fontawesome-solid-x:     | `[]`                           | A list of regular expressions that denote which subpackages should be excluded when `recursive: true` |
| `exclude-interface-regex`                              | :fontawesome-solid-x:     | `""`                           | When set along with `include-interface-regex`, then interfaces which match `include-interface-regex` but also match `exclude-interface-regex` will not be generated. If `all` is set, or if `include-interface-regex` is not set, then `exclude-interface-regex` has no effect.                        |
| `filename`                                             | :fontawesome-solid-check: | `"mocks_test.go"` | The name of the file the mock will reside in. Multiple interfaces from the same source package can be placed into the same output file.                                                                                                                             |
| `force-file-write`                                     | :fontawesome-solid-x:     | `true`                        | When set to `force-file-write: true`, mockery will forcibly overwrite any existing files. Otherwise, it will fail if the output file already exists. |
| `formatter`                                            | :fontawesome-solid-x:     | `"goimports"`                  | The formatter to use on the rendered template. Choices are: `gofmt`, `goimports`, `noop`.                                                                                                                                                            |
| `formatter-options`                                    | :fontawesome-solid-x:     | `nil`                          | Additional options for the formatter. See below.                                                            |
| `generate`                                             | :fontawesome-solid-x:     | `true`                         | Can be used to selectively enable/disable generation of specific interfaces. See [the related docs](generate-directive.md) for more details. |
| [`include-auto-generated`](include-auto-generated.md){ data-preview }                               | :fontawesome-solid-x:     | `false`                        | When set to `true`, mockery will parse files that are auto-generated. This can only be specified in the top-level config or package-level config. |
| `include-interface-regex`                              | :fontawesome-solid-x:     | `""`                           | When set, only interface names that match the expression will be generated. This setting is ignored if `all: True` is specified in the configuration. To further refine the interfaces generated, use `exclude-interface-regex`.                               |
| [`inpackage`](inpackage.md){ data-preview }            | :fontawesome-solid-x:     | `nil`                          | When set, this overrides mockery's auto-detection logic when determining if the mock file is inside or outside of the mocked interface's package. |
| `log-level`                                            | :fontawesome-solid-x:     | `"info"`                       | Set the level of the logger                                                                                                                                                                                                                          |
| `structname`                                           | :fontawesome-solid-check: | `"{{.Mock}}{{.InterfaceName}}"` | The name of the generated interface implementation.                                                                                                                                                                                                                      |
| `packages`                                             | :fontawesome-solid-x:     | `null`                         | A dictionary containing configuration describing the packages and interfaces to generate mocks for.                                                                                                                                                  |
| `pkgname`                                              | :fontawesome-solid-check: | `"{{.SrcPackageName}}"`        | The `#!go package name` given to the generated mock files.                                                                                                                                                                                           |
| `recursive`                                            | :fontawesome-solid-x:     | `false`                        | When set to `true` on a particular package, mockery will recursively search for all sub-packages and inject those packages into the config map.                                                                                                      |
| [`replace-type`](replace-type.md){ data-preview }      | :fontawesome-solid-x:     | `{}`                           | Use this parameter to specify type replacements.                                 |
| `build-tags`                                           | :fontawesome-solid-x:     | `""`                           | A space-separated list of additional build tags to load packages.                                                                                                                                                                                    |
| `require-template-schema-exists`                       | :fontawesome-solid-x:     | `true`                         | If set to `true` and the schema failed to download, mockery will fail. Otherwise, mockery will not attempt to download the file nor do any schema validation.                                                |
| `template`                                             | :fontawesome-solid-x:     | `""`                           | The template to use. The choices are defined in the [Templates](../template/) section.                                                                                                                                        |
| `template-data`                                        | :fontawesome-solid-x:     | `{}`                           | A `map[string]any` that provides arbitrary options to the template. Each template will have a different set of accepted keys. Refer to each template's documentation for more details.                                                               |
| `template-schema`                                      | :fontawesome-solid-check: | `"{{.Template}}.schema.json"`  | The URL of the JSON schema to apply to the `template-data` parameter. See the [template docs](./template/index.md#schemas){ data-preview } for more details. |


The `formatter-options` field allows formatter-specific configuration. Currently goimports is supported. For
example:

```yaml title="mockery.yaml"
---
formatter: goimports
formatter-options:
  goimports:
    local-prefix: github.com/myrepo
    tab-indent: true
```
### Template Configuration
Choosing this template will render a traditional "mockery-style" template.
```yaml
template: testify
```
