package libfoo

// Client defines public interface provided by this package
//
type Client interface {
	Method1() error
	Method2() error
	Method3() error
	Method4() error
	Method5(param1 int, param2 string) error
	Method6() error
}
