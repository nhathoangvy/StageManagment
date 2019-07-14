package main

import "github.com/casbin/casbin"

func casbinTest() {
	e := casbin.NewEnforcer("./rbac.conf", "./policy.csv")
	sub := "alice" // the user that wants to access a resource.
	obj := "data1" // the resource that is going to be accessed.
	act := "read" // the operation that the user performs on the resource.

	if e.Enforce(sub, obj, act) == true {
		// permit alice to read data1
		println("OK")
	} else {
		println("Oops")
		// deny the request, show an error
	}
}