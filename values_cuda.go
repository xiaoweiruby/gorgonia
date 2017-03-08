// +build cuda

package gorgonia

import (
	"github.com/chewxy/cu"
	"github.com/pkg/errors"
)

//  convM2V converts Memory to Value
func convM2V(m External, dev Device, mem Memory, val *Value) (err error) {
	switch mt := mem.(type) {
	case Value:
		*val = mt
	case cu.DevicePtr:
		machine := m.(CUDAMachine)
		ctxes := machine.Contexts()
		if len(ctxes) == 0 || len(ctxes) <= int(dev) {
			return errors.Errorf("Cannot convert Memory to Value when there are no CUDA contexts")
		}
		ctx := ctxes[int(dev)]
		if err = devPtrToValue(ctx, *val, mt); err != nil {
			return
		}
	}
	return nil
}

func valToDevicePointer(ctx *cu.BatchedContext, val Value) (mem cu.DevicePtr, err error) {
	// alloc:
	size := int64(val.MemSize())
	if mem, err = cu.MemAlloc(size); err != nil {
		err = errors.Wrapf(err, "Cannot get mem device pointer")
		return

	}

	// batched copy
	if ctx != nil {
		ctx.MemcpyHtoD(mem, val.Pointer(), size)
		return
	}

	// blocking copy
	if err = cu.MemcpyHtoD(mem, val.Pointer(), size); err != nil {
		err = errors.Wrapf(err, "Memcpy failed")
		return
	}
	return mem, nil
}

func devPtrToValue(ctx *cu.BatchedContext, val Value, mem cu.DevicePtr) (err error) {
	size := int64(val.MemSize())
	ptr := val.Pointer()
	if ctx != nil {
		ctx.MemcpyDtoH(ptr, mem, size)
		// ctx.DoWork()
		return nil
	}
	return cu.MemcpyDtoH(ptr, mem, size)
}
