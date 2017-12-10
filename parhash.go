package parhash

import (
	"hash"
	"runtime"
	"sync"
)

var (
	workQueue chan *hasher
)

// to parallelize work the package maintains one worker per CPU
func init() {
	numCPUs := runtime.NumCPU()
	workQueue = make(chan *hasher, numCPUs+1)

	// create a worker for each CPU
	for i := 0; i < numCPUs; i++ {
		go func() {
			for h := range workQueue {
				h.hash.Write(*h.data)
				h.wg.Done()
			}
		}()
	}
}

type hasher struct {
	hash hash.Hash

	// points to data to be written to the hash
	data *[]byte

	// share a WaitGroup with to trigger the Done()
	wg *sync.WaitGroup
}

type Parhash struct {
	wg     sync.WaitGroup
	hashes []*hasher
}

func New() *Parhash {
	return &Parhash{hashes: make([]*hasher, 0, 2)}
}

func (p *Parhash) Add(h hash.Hash) hash.Hash {
	p.hashes = append(p.hashes, &hasher{
		wg:   &p.wg,
		hash: h,
		data: nil,
	})

	return h
}

func (p *Parhash) Write(b []byte) (n int, err error) {
	for _, hasher := range p.hashes {
		p.wg.Add(1)
		hasher.data = &b
		workQueue <- hasher
	}

	p.wg.Wait()
	return len(b), nil
}

// writeSerial is only used for benchmarking to contrast
// performance differences between serial and parallel hashing
func (p *Parhash) writeSerial(b []byte) (n int, err error) {
	for _, hasher := range p.hashes {
		hasher.hash.Write(b)
	}

	return len(b), nil
}
