package main

import (
	// Import the entire framework for interracting with SDAccel from Go (including bundled verilog)
	_ "github.com/ReconfigureIO/sdaccel"

	// Use the new AXI protocol package for interracting with memory
	aximemory "github.com/ReconfigureIO/sdaccel/axi/memory"
	axiprotocol "github.com/ReconfigureIO/sdaccel/axi/protocol"
)

// function to calculate the bin for each sample
func replaceItem(m <-chan uint64,result chan<- uint64,item int,size int,input uint64)  {
	replace:=make(chan uint64,20)
	for i:=0;i<item;i++ {
		replace<-<-m
	}

	replace<-input
	<-m
	for i:=item;i<size-1;i++ {
		replace<-<-m
	}

	for i:=0;i<size ;i++  {
		result<-<-replace
	}

}
func getItem(input <-chan uint64,result chan<- uint64,size int,item int)(uint64) {

	replace:=make(chan uint64,20)

	for i:=0;i<item ;i++  {
		replace<-<-input
	}
	swap:=<-input
	replace<-swap

	for i:=item+1;i<size ;i++  {
		replace<-<-input
	}

	for i:=0;i<size ;i++  {
		result<-<-replace
	}
	return swap
}
func mergesort_iterative_cpu(arr chan uint64,size int){
	temparr:= make(chan uint64,20)
	for i:=0;i<size;i++{
		temparr<-0
	}
	var right int
	var rend int
	var i int
	var j int
	var m int

	for k:= 1; k < size; k *= 2 {
		//at each partition size, sort and merge
		for  left := 0; left + k < size; left += k*2 {
			//store the start of the right partition and its end
			right = left + k
			rend = right + k

			//if the partitions are uneven, readjust the end
			if rend > size{
				rend = size
			}
			m = left
			i = left
			j = right

			//merge
			for i < right && j < rend {


				if getItem(arr,arr,size,i) <=getItem(arr,arr,size,j) {
					replaceItem(temparr,temparr,m,size,getItem(arr,arr,size,i))
					//temparr[m] = arr[i]
					i++
				} else {
					replaceItem(temparr,temparr,m,size,getItem(arr,arr,size,j))
					//temparr[m] = arr[j]
					j++
				}
				m++
			}
			for i < right {
				//	temparr[m] = arr[i]
				replaceItem(temparr,temparr,m,size,getItem(arr,arr,size,i))
				i++
				m++
			}
			for j < rend {
				//temparr[m] = arr[j]
				replaceItem(temparr,temparr,m,size,getItem(arr,arr,size,j))
				j++
				m++
			}
			//copy from temp array into initial array
			for m = left; m < rend; m++ {
				replaceItem(arr,arr,m,size,getItem(temparr,temparr,size,m))
				//	arr[m] = temparr[m]
			}
		}
	}
}


func Top(
	// Three operands from the host. Pointers to the input data and the space for the result in shared
	// memory and the length of the input data so the FPGA knows what to expect.
	inputData uintptr,
	outputData uintptr,
	length uint32,

	// Set up channels for interacting with the shared memory
	memReadAddr chan<- axiprotocol.Addr,
	memReadData <-chan axiprotocol.ReadData,

	memWriteAddr chan<- axiprotocol.Addr,
	memWriteData chan<- axiprotocol.WriteData,
	memWriteResp <-chan axiprotocol.WriteResp) {


	// Read all of the input data into a channel
	inputChan := make(chan uint64)
	go aximemory.ReadBurstUInt64(
		memReadAddr, memReadData, true, inputData, length, inputChan)

			mergesort_iterative_cpu(inputChan,20)


	// Write the results to shared memory
	aximemory.WriteBurstUInt64(
		memWriteAddr, memWriteData, memWriteResp, true, outputData, 512, inputChan)
}
