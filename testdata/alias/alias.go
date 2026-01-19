package alias

type Foo string

type AliasNested = Alias
type Alias = Foo
type AliasString = string
type AliasStringSlice = []string

type FooStruct struct {
	Key string
}

type AliasStruct = FooStruct
type AliasStructNested = AliasStruct
type AliasStructSlice = []FooStruct
type AliasStructNestedSlice = []AliasStructNested

// RemappedAlias should be manually remapped to "string" in the test settings.
type RemappedAlias = FooStruct

type UseAliasedType struct {
	Field1 RemappedAlias
}
