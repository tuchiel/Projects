package main

import (
	"Projects/myLoger"
	"Projects/myTimer"
	"fmt"
	"time"
)

var loger = logg.DefaultModuleLogger

func myHandler(x interface{}) bool {
	*(x.(*uint64))++
	loger.Debug("Stored int is : %d\n", *(x.(*uint64)))
	return true
}

func oneShotHandler(x interface{}) bool {
	*(x.(*chan bool)) <- true
	loger.Debug("Unblocking main program")
	return true
}

func main() {
	mgr := &timex.Manager{}
	i := uint64(0)
	ch := make(chan bool)
	t := mgr.CreateLimitedPeriodicTimer(5, time.Second, myHandler, &i)
	t2 := mgr.CreatePeriodicTimer(time.Second, myHandler, &i)
	t3 := mgr.CreateOneTimeTimer(time.Second*7, oneShotHandler, &ch)
	mgr.PrintTimers()
	mgr.Start(t)
	mgr.Start(t2)
	mgr.Start(t3)
	time.Sleep(time.Second * 2)
	mgr.PrintTimers()

	<-ch
	mgr.Stop(t)
	mgr.Stop(t2)
	mgr.Stop(t3)
	mgr.PrintTimers()

	fmt.Println(i)
}
