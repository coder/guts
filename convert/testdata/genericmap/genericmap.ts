// Code generated by 'gots'. DO NOT EDIT.

// From codersdk/genericmap.go
interface Buzz {
    bazz: string;
}

// From codersdk/genericmap.go
type Custom = Foo | Buzz;

// From codersdk/genericmap.go
interface Foo {
    bar: string;
}

// From codersdk/genericmap.go
interface FooBuzz<R extends Custom> {
    something: R[];
}
