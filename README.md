# rdmanl

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/nickgarlis/rdmanl)
[![GitHub](https://img.shields.io/github/license/nickgarlis/rdmanl)](https://img.shields.io/github/license/nickgarlis/rdmanl)

A pure Go client for RDMA netlink, the kernel interface behind the `rdma` tool
from iproute2. It lists RDMA devices, ports, resources and counters, creates
and deletes software devices like rxe, and watches for device events.

## Example

```go
package main

import (
	"fmt"

	"github.com/nickgarlis/rdmanl"
)

func main() {
	conn, err := rdmanl.Open(nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	devs, err := conn.Devices()
	if err != nil {
		panic(err)
	}

	for _, d := range devs {
		ports, err := conn.Ports(rdmanl.PortReq{DevIndex: d.Index})
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s: %d ports\n", d.Name, len(ports))
	}
}
```

## License

MIT, see [LICENSE](https://github.com/nickgarlis/rdmanl/blob/main/LICENSE).
