package main

import (
	"time"
	"fmt"
)

func main (){
	//
	//

	//func() {
	//
	//	for i:=0 ; i<10 ; i++ {
	//
	//		d := time.Duration(i) * time.Second
	//		t := time.NewTimer(d)
	//		//time.NewTicker()
	//		fmt.Println("timer check point 0 @",i,<-t.C)
	//
	//		go func(c <-chan time.Time) {
	//			fmt.Println("DO SOMETHING",<-c)
	//			//fmt.Println("timer check point 1 %s",c)
	//		}(t.C)
	//		fmt.Println("DO SOMETHING outer")
	//
	//	}
	//
	//
	//
	//	fmt.Println("schedule continue")
	//
	//	time.Sleep(2*time.Second)
	//
	//	fmt.Println("22schedule continue")
	//
	//	//y:=t.Stop()
	//	//	if !y {
	//	//		<-t.C
	//	//		}
	//	//b := t.Reset(d)
	//	//p :=<-t.C
	//	//fmt.Println("timer check point 2 %d %s",b,p)
	//
	//	return
	//}()
	//
	//
	//return


	sema := make (chan struct{},1)
	//sema<- struct {}{}
	c1 := make(chan string, 1)
	c1<-"s"
	//<-c1
	fmt.Println(" get chan")
	//c1<-"s"



	go func() {
		time.Sleep(time.Second * 2)
		//c1 <- "result 1"
		fmt.Println(" in func ")
		sema <- struct{}{}
	}()
	fmt.Println(" in main ")

	<-sema
	fmt.Println("xxxxxxx")
//	fff := func (x,unit float64) float64 {
//		return (math.Ceil(x/unit)) * unit
//	}
//
//	fff(0.34,0.5)
//	fff(2.346,0.5)
//	fff(0.31,0.5)
//	fff(0.49,0.5)
//
//
//
//	x := math.Ceil(1.556)
//fmt.Println(x)
//	layout := "15:04:05"
//	lt := "26:34:55"
//	t,err := time.Parse(layout,lt)
//	fmt.Println(t,err)


	//sign := make(chan int)
	//
	//i := 10
	//d := 3 * time.Second
	//
	//createTime := func() *time.Timer{
	//	t := time.NewTimer(d)
	//	return t
	//}
	//
	//process := func(t *time.Timer) {
	//	expire := <-t.C
	//	fmt.Printf("Expiration time: %v.\n", expire)
	//	t.Reset(d)
	//}
	//
	//go func (){
	//
	//	t := createTime()
	//	now := time.Now()
	//	fmt.Printf("Now time : %v.\n", now)
	//	process(t)
	//	fmt.Println("xxxxxx")
	//	i--
	//	if i ==0 {
	//		sign<-1
	//	}
	//
	//}()
	//fmt.Printf("un buff??")
	//
	//<-sign

	//channelTest(make(<-chan time.Time,1))
}

var intChan chan int

func channelTest(x <-chan time.Time) {
	t := time.NewTimer(2 * time.Second)
	fmt.Println("time %s", t)
	//x2 := <-t.C

	fmt.Println("time 1111111 %s", <-t.C)

}