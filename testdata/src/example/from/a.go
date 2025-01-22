package from

import (
	_ "example/to"           // want "invalid dependency: example/to"
	_ "example/to/exception" // ok
)
