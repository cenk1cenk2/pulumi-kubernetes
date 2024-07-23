package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantPath  string
		wantValue string
		wantErr   string
	}{
		{
			name:    "empty expression",
			expr:    "",
			wantErr: "non-empty",
		},
		{
			name:    "missing prefix",
			expr:    "{.foo}",
			wantErr: "jsonpath=",
		},
		{
			name:    "quoted key with value",
			expr:    "jsonpath='{.status.phase}'=Running",
			wantErr: "omit shell quotes",
		},
		{
			name:    "missing value",
			expr:    "jsonpath={.metadata.name}=",
			wantErr: "{.metadata.name}= requires a value",
		},
		{
			name:    "invalid expression with repeated =",
			expr:    "jsonpath={.metadata.name}='test=wrong'",
			wantErr: "format should be {.path}=value or {.path}",
		},
		{
			name:    "complex expressions are not supported",
			expr:    "jsonpath={.status.conditions[?(@.type==\"Failed\"||@.type==\"Complete\")].status}=True",
			wantErr: "unrecognized character",
		},
		{
			name:     "key with any value",
			expr:     "jsonpath={.foo}",
			wantPath: "{.foo}",
		},
		{
			name:      "key with value",
			expr:      "jsonpath={.foo}=bar",
			wantPath:  "{.foo}",
			wantValue: "bar",
		},
		{
			name:      "preserve ==",
			expr:      `jsonpath={.status.containerStatuses[?(@.name=="foobar")].ready}=True`,
			wantPath:  `{.status.containerStatuses[?(@.name=="foobar")].ready}`,
			wantValue: "True",
		},
		{
			name:     "padded brackets",
			expr:     "jsonpath={ .webhooks[].clientConfig.caBundle }",
			wantPath: `{ .webhooks[].clientConfig.caBundle }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := Parse(tt.expr)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, instance.path)
			assert.Equal(t, tt.wantValue, instance.value)
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		uns       *unstructured.Unstructured
		wantMatch bool
		// wantInfo  string
		wantErr string
	}{
		{
			name:      "no match",
			expr:      "jsonpath={.foo}",
			uns:       &unstructured.Unstructured{Object: map[string]any{}},
			wantMatch: false,
		},
		{
			name: "key exists",
			expr: "jsonpath={ .foo }",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": nil,
			}},
			wantMatch: true,
		},
		{
			name: "key exists with non-primitive value",
			expr: "jsonpath={.foo}",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": []string{"boo"},
			}},
			wantMatch: true,
		},
		{
			name: "value matches",
			expr: "jsonpath={.foo}=bar",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": "bar",
			}},
			wantMatch: true,
		},
		{
			name: "value mismatch",
			expr: "jsonpath={.foo}=bar",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": "baz",
			}},
			wantMatch: false,
		},
		{
			name: "value match against some array element",
			expr: "jsonpath={.foo[*].bar}=baz",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": []any{
					map[string]any{
						"ignored": "true",
					},
					map[string]any{
						"bar": "baz",
					},
					map[string]any{
						"something else": "true",
					},
				},
			}},
			wantMatch: true,
		},
		{
			name: "value match against specific array element",
			expr: "jsonpath={.foo[1].bar}=baz",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": []any{
					map[string]any{
						"bar": "not-baz",
					},
					map[string]any{
						"bar": "baz",
					},
				},
			}},
			wantMatch: true,
		},
		{
			name: "value mismatch against specific array element",
			expr: "jsonpath={.foo[0].bar}=baz",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": []any{
					map[string]any{
						"bar": "not-baz",
					},
					map[string]any{
						"bar": "baz",
					},
				},
			}},
			wantMatch: false,
		},
		{
			name: "value match against non-primitive value",
			expr: "jsonpath={.foo}=bar",
			uns: &unstructured.Unstructured{Object: map[string]any{
				"foo": []any{"bar"},
			}},
			wantErr: "has a non-primitive value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, err := Parse(tt.expr)
			require.NoError(t, err)

			actual, _, err := i.Matches(tt.uns)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, actual)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		given Parsed
		want  string
	}{
		{
			given: Parsed{path: "{.foo}"},
			want:  "{.foo}",
		},
		{
			given: Parsed{path: "{.foo}", value: "1"},
			want:  "{.foo}=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.given.String())
		})
	}
}
