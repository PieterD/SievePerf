package main

import "time"
import "fmt"

// Nicked from the Golang heap library because I needed down.
func up(list [][2]uint64, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || (list[i][0] < list[j][0]) {
			break
		}
		list[i], list[j] = list[j], list[i]
		j = i
	}
}

func down(list [][2]uint64, i int) {
	n := len(list)
	for {
		j1 := 2*i + 1
		if j1 >= n {
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !(list[j1][0] < list[j2][0]) {
			j = j2 // right child
		}
		if list[i][0] < list[j][0] {
			break
		}
		list[i], list[j] = list[j], list[i]
		i = j
	}
}

func push(list [][2]uint64, a,b uint64) [][2]uint64 {
	list = append(list, [2]uint64{a,b})
	up(list, len(list)-1)
	return list
}

// The original.
func sieve0(max uint64) []uint64 {
	var list []uint64
	var D = make(map[uint64][]uint64)
	var q uint64 = 2
	for q <= max {
		// Check if q is in the composite list.
		l,ok := D[q]
		if !ok {
			// If it isn't, it's prime.
			list = append(list, q)
			// Also add the square to the compisite list.
			D[q*q] = []uint64{q}
		} else {
			// q is in the composite list
			for _,p := range l {
				// The D[q] contains a list of primes for whom q is a prime
				// power. Go through them all and add the next prime power
				// to the list.
				D[p+q] = append(D[p+q], p)
			}
			// Remove D[q]; we won't visit it again.
			delete(D, q)
		}
		// Next q
		q++
	}
	return list
}

// Now avoids all even numbers.
func sieve1(max uint64) []uint64 {
	// Start with two primes already in the list.
	var list = []uint64{2,3}
	var D = make(map[uint64][]uint64)
	// Since 3 is on the list, add its next prime power to D. 2 doesn't
	// have to be added because all its prime powers are even, and
	// we're skipping even numbers.
	D[9] = []uint64{3}
	// Start at 5, the next prime after 3.
	var q uint64 = 5
	for q <= max {
		l,ok := D[q]
		if !ok {
			list = append(list, q)
			D[q*q] = []uint64{q}
		} else {
			for _,p := range l {
				// Whenever we add a prime power to D, we skip one because we're
				// only checking uneven numbers. This means q is uneven, and so
				// is p (because the only even prime is 2). Adding two uneven
				// numbers results in an even number, which we want to skip;
				// therefore we add p twice.
				nv := p*2+q
				D[nv] = append(D[nv], p)
			}
			delete(D, q)
		}
		// The next q is q+2.
		q += 2
	}
	return list
}

// Now also avoids all multiples of 3.
func sieve2(max uint64) []uint64 {
	var list = []uint64{2,3}
	var D = make(map[uint64][]uint64)
	// Since we're skipping multiples of 3, don't add prime powers of 3 to D.
	//D[9] = []uint64{3}
	var q uint64 = 5
	// threeable flips between 0 and 1 every cycle.
	var threeable uint64 = 0
	for q <= max {
		l,ok := D[q]
		if !ok {
			list = append(list, q)
			D[q*q] = []uint64{q}
		} else {
			for _,p := range l {
				// Skip even numbers like before
				nv := p*2+q
				if nv%3 == 0 {
					// Now, if nv is divisible by 3, add another
					// 2p to q.
					nv += p*2
				}
				D[nv] = append(D[nv], p)
			}
			delete(D, q)
		}
		// This time, q is incremented by 2 and 4 alternating.
		// The reason for this is this:
		// 1  2  3  4  5  6  7  8  9  10 11 12 13 14 15 16 17 18 19 20 21 22 23
		//    *     *     *     *     *     *     *     *     *     *     *
		//       *        *        *        *        *        *        *
		// 4           2     4           2     4           2     4           2
		// Starting from 1, the next not-multiple-of-2-or-3 is 1+4=5. The next
		// is 5+2, 7+4, 11+2, 13+4, 17+2, 19+4, 23+2.
		q += 2 + 2*threeable
		// flip threeable.
		threeable = 1-threeable
	}
	return list
}

// This one uses a heap instead of a hash table.
func sieve3(max uint64) []uint64 {
	var list = []uint64{2,3,5}
	// We start from 7 this time, so add 5^2 to the prime power list.
	// We do this so we don't have to worry about the special case of an
	// empty heap.
	var D = [][2]uint64{{25, 5}}
	var q uint64 = 7
	// Since we start at 7, threeable starts at 1 so we increment by 4
	// the first time.
	var threeable uint64 = 1
	for q <= max {
		l := D[0]
		if l[0] != q {
			list = append(list, q)
			D = push(D, q*q, q)
		} else {
			// The primes that make up the prime power q are no longer in a
			// list, but individual elements in a priority queue. That means
			// We check if the head == q until it isn't.
			for D[0][0] == q {
				p := D[0][1]
				nv := p*2+q
				if nv%3 == 0 {
					nv += p*2
				}
				D[0][0] = nv
				down(D, 0)
				// The above is still the same as before, it just looks a bit
				// worse.
			}
		}
		q += 2 + 2*threeable
		threeable = 1-threeable
	}
	return list
}

// This one uses a wheel sieve to avoid all numbers divisible by the first
// num primes.
func sivbuild(num int) (uint64, uint64, []uint64, []byte) {
	// The size of the wheel is exponential. Anything more than 10 primes
	// will create an enormous wheel; try 7.
	var startprimes = []uint64{2,3,5,7,11,13,17,19,23,29}
	n1 := startprimes[num]
	n2 := startprimes[num+1]
	sl := startprimes[:num+1]
	startprimes = startprimes[:num]

  // Multiply our starting primes.
	var product int = 1
	for i := range startprimes {
		product *= int(startprimes[i])
	}

  // We start with an array with one element for every number less than
	// product, initialized to 0.
	var wheel = make([]byte, product+1)
	// Set the element at every prime power for every starting prime to 1.
	for _,p := range startprimes {
		for i:=int(p); i<=product; i+=int(p) {
			wheel[i] = 1
		}
		/*
		for i:=int(2*p); i<=product; i+=int(p) {
			wheel[i] = 1
		}
		wheel[p] = 1
		*/
	}

  // The increment to the next possible-prime from the current maybe-prime
	// is equal to the number of ones that follow our current maybe-prime,
	// plus one. To get this, traverse the array in reverse, adding up ones
	// until we find a zero.
	// Halve the size of the array by ignoring even numbers.
	var c int = 1
	var wheel2 = make([]byte, product/2)
	for i:=product; i>0; i-- {
		if wheel[i] == 1 {
			c++
		} else {
			if c > 255 {
				panic("OH GOD c > 255")
			}
			wheel2[i/2] = byte(c)
			c=1
		}
	}

	return n2, n1, sl, wheel2
}
// Look up the increment to reach the next possible-prime from the given
// maybe-prime.
func lookup(wheel []byte, num uint64) uint64 {
	return uint64(wheel[(num%uint64(len(wheel)*2))/2])
}
// Move some horribleness from the inner loop.
func updateheap(D [][2]uint64, wheel []byte, q uint64) {
	p := D[0][1]
	nv := D[0][0]
	m := nv/p
	// As before, we add p the same number of times as we increment q.
	// That way we don't have to clean up all the prime powers less than
	// q every time we cycle.
	nv += p*lookup(wheel, m)
	D[0][0] = nv
	down(D, 0)
}
func sieve4(max uint64, init int) []uint64 {
	q, b, list, wheel := sivbuild(init)
	var D = [][2]uint64{{b*b, b}}

	for q <= max {
		if D[0][0] != q {
			list = append(list, q)
			D = push(D, q*q, q)
		} else {
			for D[0][0] == q {
				updateheap(D, wheel, q)
			}
		}
		q += lookup(wheel, q)
	}
	return list
}

// This one is basically sieve4, but with a channel interface instead
// of one big list.
func sieve5(max uint64, init int) (list []uint64) {
	s := sieve5run(init)
	for {
		p := s.Get()
		if p > max {
			break
		}
		list = append(list, p)
	}
	s.Close()
	return
}
type Sieve struct {
	primechan <-chan uint64
	quitchan chan<- bool
}
func (s *Sieve) Close() {
	close(s.quitchan)
	for <-s.primechan != 0 {
		;
	}
}
func (s *Sieve) Get() uint64 {
	return <-s.primechan
}
func sivbuild5(num int, ch chan<- uint64) (uint64, uint64, []byte) {
	var startprimes = []int{2,3,5,7,11,13,17,19,23,29}
	n1 := startprimes[num]
	n2 := startprimes[num+1]
	startprimes = startprimes[:num]


	var product int = 1
	for i := range startprimes {
		ch <- uint64(startprimes[i])
		product *= startprimes[i]
	}
	ch <- uint64(n1)

	var wheel = make([]byte, product+1)
	for _,p := range startprimes {
		for i:=2*p; i<=product; i+=p {
			wheel[i] = 1
		}
		wheel[p] = 1
	}

	var c = 1
	var wheel2 = make([]byte, product/2)
	for i:=product; i>0; i-- {
		if wheel[i] == 1 {
			c++
		} else {
			if c > 255 {
				panic("OH GOD c > 255")
			}
			wheel2[i/2] = byte(c)
			c=1
		}
	}

	return uint64(n2), uint64(n1), wheel2
}
func sieve5run(init int) (s *Sieve) {
	s = new(Sieve)
	mprimechan := make(chan uint64, 10)
	mquitchan := make(chan bool)
	s.primechan = mprimechan
	s.quitchan = mquitchan
	var primechan chan<- uint64 = mprimechan
	var quitchan <-chan bool = mquitchan
	go func(){
		q, b, wheel := sivbuild5(init, primechan)
		var D = [][2]uint64{{b*b, b}}
		cont := true
		for cont {
			if D[0][0] != q {
				select {
					case primechan <- q:
					case <-quitchan:
						close(primechan)
						cont = false
				}
				D = push(D, q*q, q)
			} else {
				for D[0][0] == q {
					updateheap(D, wheel, q)
				}
			}
			q += lookup(wheel, q)
		}
	}()
	return
}

func main() {
	const n = 50000000
	//const n = 50000000
	ts := time.Now().UnixNano()
	l0 := sieve0(n)
	t0 := time.Now().UnixNano()
	fmt.Printf("Sieve0: %f\n", float64(t0-ts)/1000000000)
	l1 := sieve1(n)
	t1 := time.Now().UnixNano()
	fmt.Printf("Sieve1: %f\n", float64(t1-t0)/1000000000)
	l2 := sieve2(n)
	t2 := time.Now().UnixNano()
	fmt.Printf("Sieve2: %f\n", float64(t2-t1)/1000000000)
	l3 := sieve3(n)
	t3 := time.Now().UnixNano()
	fmt.Printf("Sieve3: %f\n", float64(t3-t2)/1000000000)
	l4 := sieve4(n, 7)
	t4 := time.Now().UnixNano()
	fmt.Printf("Sieve4: %f\n", float64(t4-t3)/1000000000)
	l5 := sieve5(n, 7)
	t5 := time.Now().UnixNano()
	fmt.Printf("Sieve5: %f\n", float64(t5-t4)/1000000000)

	if len(l0) != len(l1) {
		fmt.Printf("%d != %d\n", len(l0), len(l1))
		panic("LENGTH 1 DONT MATCH!")
	}
	if len(l1) != len(l2) {
		fmt.Printf("%d != %d\n", len(l1), len(l2))
		panic("LENGTH 2 DONT MATCH!")
	}
	if len(l2) != len(l3) {
		fmt.Printf("%d != %d\n", len(l2), len(l3))
		panic("LENGTH 3 DONT MATCH!")
	}
	if len(l3) != len(l4) {
		fmt.Printf("%d != %d\n", len(l3), len(l4))
		panic("LENGTH 4 DONT MATCH!")
	}
	if len(l4) != len(l5) {
		fmt.Printf("%d != %d\n", len(l4), len(l5))
		panic("LENGTH 5 DONT MATCH!")
	}
	for i := range l4 {
		if l0[i] != l1[i] {
			panic("HALP 1!")
		}
		if l1[i] != l2[i] {
			panic("HALP 2!")
		}
		if l2[i] != l3[i] {
			panic("HALP 3!")
		}
		if l3[i] != l4[i] {
			panic("HALP 4!")
		}
		if l4[i] != l5[i] {
			panic("HALP 5!")
		}
	}
	te := time.Now().UnixNano()
	fmt.Printf("Prime check: %f\n", float64(te-t4)/1000000000)
}

