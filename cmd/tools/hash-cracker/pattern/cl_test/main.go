package main

import (
	"log"
	"slices"
	"unsafe"

	cl "github.com/CyberChainXyz/go-opencl"
)

// TestRunner is a test function that tests the functionality of the OpenCL runner.
func main() {
	// Step 1: Get OpenCL device information
	info, _ := cl.Info()

	if len(info.Platforms) < 1 {
		log.Fatal("No OpenCL Devices")
	}
	if len(info.Platforms[0].Devices) < 1 {
		log.Fatal("No OpenCL Devices")
	}

	// Step 2: Initialize the OpenCL runner
	device := info.Platforms[0].Devices[0]
	runner, err := device.InitRunner()
	if err != nil {
		log.Fatal("InitRunner err:", err)
	}
	defer runner.Free()

	// Step 3: Compile the OpenCL kernels
	code := `__kernel void helloworld(__global int* in, __global int* out)
		 {
			 int num = get_global_id(0);
			 out[num] = in[num] * in[num];
		 }`
	codes := []string{code}
	kernelNameList := []string{"helloworld"}
	err = runner.CompileKernels(codes, kernelNameList, "")
	if err != nil {
		log.Fatal("CompileKernels err:", err)
	}

	// Step 4: Create kernel parameters
	/* buffer 1 param */
	input := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	itemSize := int(unsafe.Sizeof(input[0]))
	itemCount := len(input)

	input_buf, err := cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, input)
	if err != nil {
		log.Fatal("CreateBuffer err:", err)
	}
	/* buffer 2 param */
	output_buf, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, itemCount*itemSize)
	if err != nil {
		log.Fatal("CreateEmptyBuffer err:", err)
	}

	// Step 5: Run the OpenCL kernel
	err = runner.RunKernel("helloworld", 1, nil, []uint64{uint64(itemCount)}, nil, []cl.KernelParam{
		cl.BufferParam(input_buf),
		cl.BufferParam(output_buf),
	}, true)
	if err != nil {
		log.Fatal("RunKernel err:", err)
	}

	// Step 6: Read the output buffer
	result := make([]int32, itemCount)
	err = cl.ReadBuffer(runner, 0, output_buf, result)
	if err != nil {
		log.Fatal("ReadBuffer err:", err)
	}
	log.Printf("Result: %v", result)

	// Step 7: Check the result
	expected_result := make([]int32, itemCount)
	for i, v := range input {
		expected_result[i] = v * v
	}
	if !slices.Equal(result, expected_result) {
		log.Fatal("result error:", result)
	}

}
