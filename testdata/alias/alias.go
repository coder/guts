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
