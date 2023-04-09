;; Copyright 2023 The Wa Authors. All rights reserved.

(func $$waGetStackPtr (result i32)
	(global.get $__stack_ptr)
)

(func $$waSetStackPtr (param $sp i32)
	local.get $sp
	global.set $__stack_ptr
)

(func $$waStackAlloc (param $size i32) (result i32)
	;; $__stack_ptr -= $size
	(global.set $__stack_ptr (i32.sub (global.get $__stack_ptr) (local.get  $size)))
	;; return $__stack_ptr
	(return (global.get $__stack_ptr))
)

(func $$waStackFree (param $size i32)
	;; $__stack_ptr += $size
	(global.set $__stack_ptr (i32.add (global.get $__stack_ptr) (local.get $size)))
)

(func $$waHeapBase(result i32)
	global.get $__heap_base
)

(func $$waHeapMax(result i32)
	global.get $__heap_max
)

(func $$waHeapAlloc (param $nbytes i32) (result i32) ;;result = ptr
	(local $ptr i32)

	local.get $nbytes
	i32.eqz
	if
		i32.const 0
		return
	end

	local.get $nbytes
	i32.const 7
	i32.add
	i32.const 8
	i32.div_u
	i32.const 8
	i32.mul
	local.set $nbytes

	local.get $nbytes
	call $runtime.malloc
	local.set $ptr

	loop $zero
		local.get $nbytes
		i32.const 8
		i32.sub
		local.tee $nbytes
		local.get $ptr
		i32.add

		i64.const 0
		i64.store

		local.get $nbytes
		if
			br $zero
		end
	end ;;loop $zero

	local.get $ptr

	;;Todo
	;; global.get $__heap_base

	;; global.get $__heap_base
	;; call $$runtime.waPrintI32
	;; i32.const 32
	;; call $$runtime.waPrintRune
	;; local.get $size
	;; call $$runtime.waPrintI32
	;; i32.const 10
	;; call $$runtime.waPrintRune

	;; global.get $__heap_base
	;; global.get $__heap_base
	;; local.get $nbytes
	;; i32.const 7
	;; i32.add
	;; i32.const 8
	;; i32.div_u
	;; i32.const 8
	;; i32.mul
	;; i32.add
	;; global.set $__heap_base
	;; call $$wa.RT.DupWatStack
	;; call $$runtime.waPrintI32
	;; i32.const 10
	;; call $$runtime.waPrintRune
)

(func $$waHeapFree (param $ptr i32)
	local.get $ptr
	call $runtime.free

	;;Todo
	;; i32.const 126
	;; call $$runtime.waPrintRune
	;; local.get $ptr
	;; call $$runtime.waPrintI32
	;; i32.const 10
	;; call $$runtime.waPrintRune
)

(func $$wa.RT.Block.Init (param $ptr i32) (param $item_count i32) (param $release_func i32) (param $item_size i32) (result i32) ;;result = ptr
	local.get $ptr

	local.get $ptr
	if
		local.get $ptr
		i32.const 1
		i32.store offset=0 align=1

		local.get $ptr
		local.get $item_count
		i32.store offset=4 align=1

		local.get $ptr
		local.get $release_func
		i32.store offset=8 align=1

		local.get $ptr
		local.get $item_size
		i32.store offset=12 align=1
	end
)

(func $$wa.RT.DupWatStack (param $p i32) (result i32) (result i32) ;;result0 = result1 = p
	local.get $p
	local.get $p
)

(func $$wa.RT.Block.Retain (param $ptr i32) (result i32) ;;result = ptr
	local.get $ptr

	local.get $ptr
	if
		local.get $ptr
		local.get $ptr
		i32.load offset=0 align=1
		i32.const 1
		i32.add
		i32.store offset=0 align=1
	end
)

(func $$wa.RT.Block.Release (param $ptr i32)
	;;Todo
	(local $ref_count i32)
	(local $item_count i32)
	(local $free_func i32)
	(local $item_size i32)
	(local $data_ptr i32)

	local.get $ptr
	i32.const 0
	i32.eq
	if
		return
	end

	local.get $ptr
	i32.load offset=0 align=1
	i32.const 1
	i32.sub
	local.set $ref_count

	local.get $ref_count
	if
		local.get $ptr
		local.get $ref_count
		i32.store offset=0 align=1

	else  ;;ref_count == 0
		local.get $ptr
		i32.load offset=8 align=1
		local.set $free_func

		local.get $free_func
		if  ;;free_func != 0
			local.get $ptr
			i32.load offset=4 align=1
			local.set $item_count

			local.get $item_count
			if  ;;item_count > 0
				local.get $ptr
				i32.load offset=12 align=1
				local.set $item_size

				local.get $ptr
				i32.const 16
				i32.add
				local.set $data_ptr

				loop $free_next
					;; onFree(data_ptr)
					local.get $data_ptr
					local.get $free_func
					call_indirect (type $$onFree)

					;; item_count--
					local.get $item_count
					i32.const 1
					i32.sub
					local.set $item_count

					local.get $item_count
					if  ;;while item_count>0
						;; data_ptr += item_size
						local.get $data_ptr
						local.get $item_size
						i32.add
						local.set $data_ptr

						br $free_next  ;;continue
					end  ;;while item_count>0
				end  ;;loop $free_next
			end  ;;if item_count > 0
		end  ;;free_func != 0

		local.get $ptr
		call $$waHeapFree
	end  ;;ref_count == 0
)

(func $$wa.RT.i32_ref_to_ptr (param $b i32) (param $d i32) (result i32) ;;result = ptr
  local.get $d
)

(func $$wa.RT.slice_to_ptr (param $b i32) (param $d i32) (param $l i32) (param $c i32) (result i32) ;;result = ptr
  local.get $d
)