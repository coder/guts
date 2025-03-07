// Code generated by 'guts'. DO NOT EDIT.

// From genericstruct/genericstruct.go
export interface Baz<S extends Foo<string, string>, I extends Comparable, X extends Foo<I, I>> {
    readonly A: S;
    readonly B: X;
    readonly C: I;
}

export type Comparable = string | number | boolean;

// From genericstruct/genericstruct.go
export interface Foo<A extends Comparable, B extends any> {
    readonly FA: A;
    readonly FB: B;
}
