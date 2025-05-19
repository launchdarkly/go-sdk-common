package ldcommon

// Option represents an optional value: every Option is either Some and contains a value, or None, and does not.
type Option[T any] struct {
	value     T
	isDefined bool
}

// Some creates an Option with a value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, isDefined: true}
}

// None creates an Option with no value.
func None[T any]() Option[T] {
	return Option[T]{isDefined: false}
}

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool {
	return o.isDefined
}

// IsNone returns true if the Option does not contain a value.
func (o Option[T]) IsNone() bool {
	return !o.isDefined
}

// Unwrap returns the contained value, or panics if the Option is None.
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called Unwrap on a None Option")
	}
	return o.value
}

// UnwrapOr returns the contained value, or a provided default if the Option is None.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.IsNone() {
		return defaultValue
	}
	return o.value
}

// Map applies a function to the contained value (if any), and returns a new Option containing the result.
func Map[T any, U any](o Option[T], f func(T) U) Option[U] {
	if o.IsNone() {
		return None[U]()
	}
	return Some(f(o.value))
}

// OrElse returns the Option if it contains a value, or calls the provided function and returns its result otherwise.
func (o Option[T]) OrElse(f func() Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}
	return f()
}
