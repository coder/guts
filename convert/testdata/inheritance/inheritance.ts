// Code generated by 'gots'. DO NOT EDIT.

// From codersdk/inheritance.go
interface Bar {
    BarField: number;
}

// From codersdk/inheritance.go
interface Foo extends Bar, GenBar<string> {
}

type Comparable = string | number | boolean;

// From codersdk/inheritance.go
interface GenBar<T extends Comparable> {
    GenBarField: T;
}
