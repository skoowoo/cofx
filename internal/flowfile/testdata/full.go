package testdata

const OneFlowfile string = `
load file:///root/action1
load http://url/action2
load https://github.com/path/action3
load action4

// 变量设置
var v=1
 
set @action1
	// 输入（参数，可以一行表达，也可以多行）
	input k1=v1 k2=v2
	input k3=v3
	input k=$v

	// 输出设置, output 始终是一个 kv map 对象，打印出来是一个 json 字符串
	output out1

	loop 5 2
end

set @action2
	input k=$v

	// 传递 output 给 action (使用 input)
	input action1_out=$out1
end

// 串行执行
run @action1 m=n
run @action2
run @action3
`
