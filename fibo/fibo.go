package fibo

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

func fibonaci(input uint64) uint64 {
	if input == 0 {
		return 0
	}
	if input == 1 {
		return 1
	}
	return fibonaci(input-1) + fibonaci(input-2)
}

func assignAdd(res *uint64, op1 *uint64, op2 *uint64) {
	*res = *op1 + *op2
}

func Fibonaci2(input uint64, res *[]uint64) {
	switch input {
	case 0:
		(*res)[input] = 0
	case 1:
		(*res)[input] = 1
	default:
		defer assignAdd(&((*res)[input]), &((*res)[input-1]), &((*res)[input-2]))
		Fibonaci2(input-1, res)
		//fibonaci2(input-2, res)
	}
}

func fibAsyncWrapper(item uint64, res *[]uint64, m *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	temp := fibonaci(item)
	m.Lock()
	(*res)[item] = temp
	m.Unlock()
}

func increase(x *uint64) {
	*x++
}

func snow() {
	for x := 0; x < 120; {
		fmt.Print("*")
		x += rand.Intn(5)
		lim := rand.Intn(5)
		for i := 0; i <= lim; i++ {
			fmt.Print(" ")
		}
	}
}

func compute() {
	//var mutex = sync.Mutex{}
	start := time.Now()
	rand.Seed(time.Now().UnixNano())
	sequenceEnd := uint64(rand.Intn(80))
	/*for i := uint64(0); i <= sequenceEnd; i++ {
		fmt.Println(fibonaci(i))
	}
	syncDuration := (time.Now().UnixNano() - start.UnixNano())

	result := make([]uint64, int(sequenceEnd+1))
	var wg sync.WaitGroup

	start = time.Now()
	for i := uint64(0); i <= sequenceEnd; i++ {
		wg.Add(1)
		go fibAsyncWrapper(i, &result, &mutex, &wg)
	}
	fmt.Printf("Wating for aysnc to finish : %d\n", time.Now().UnixNano())
	wg.Wait()
	asyncDuration := (time.Now().UnixNano() - start.UnixNano())
	fmt.Printf("aysnc done at : %v\n", time.Now().UnixNano())
	fmt.Println(result)
	fmt.Printf("Async computation took : %v\n", asyncDuration)
	fmt.Printf("Async was : %v times faster\n", (float64(syncDuration) / float64(asyncDuration)))*/

	start = time.Now()

	result2 := make([]uint64, int(sequenceEnd+1))

	Fibonaci2(sequenceEnd, &result2)
	heuristicDuration := (time.Now().UnixNano() - start.UnixNano())
	//fmt.Printf("Sync  computation took : %d\n", syncDuration)
	//fmt.Printf("Async computation took : %v\n", asyncDuration)
	fmt.Printf("Heur  computation took : %v\n", heuristicDuration)
	//fmt.Printf("Heuristic was : %v times faster\n", (float64(asyncDuration) / float64(heuristicDuration)))
	//fmt.Printf("Async was : %v times faster\n", (float64(syncDuration) / float64(asyncDuration)))
	fmt.Println(result2)
	for i := 0; i < 40; i++ {
		time.Sleep(100000000)
		snow()
		fmt.Println()
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("****  ****      **** *****  ****    ******      ****   **** ")
	time.Sleep(100000000)
	fmt.Println("****   ****    ****  ***** ***     ********    ****** ******    ")
	time.Sleep(100000000)
	fmt.Println("****    ****  ****   ********     ****   ****   ***********   ")
	time.Sleep(100000000)
	fmt.Println("****     ********    ***** ***   *************   *********  ")
	time.Sleep(100000000)
	fmt.Println("****      *****      *****  **** *****   *****    ******* ")
	time.Sleep(100000000)
	fmt.Println("                                                   *****      ")
	fmt.Println("                                                    ***     ")
	fmt.Println("                                                     *      ")

}
