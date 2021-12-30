# clclyLoop
动态循环定时任务实现

## 工具
* golang(1.7+)

## 引入地址
https://github.com/deadspacewii/clclyLoop

## 使用描述
- 使用NewTimeTask方法发起一个实例，eventOption参数使用golang的选项模式实现不定长参数初始化
- Add函数实现具体执行函数(包含返回值)
- Update函数具体实现定时任务的中间修改

```go
// define hook function
func hook1(attr []interface{}) {
	if len(attr) != 3 {
		return
	}
	t := attr[0].(*TimeTask)
	left := attr[1].(string)
	right := attr[2].(string)
	if left == right {
		t.ModifyTime(3 * time.Second)
		t.Start()
	}
}

func hook2(attr []interface{}) {
	if len(attr) != 3 {
		return
	}
	t := attr[0].(*TimeTask)
	left := attr[1].(string)
	right := attr[2].(string)
	if left == right {
		t.ModifyTime(1 * time.Second)
		t.Start()
	}
}

func hook3(attr []interface{}) {
	if len(attr) != 3 {
		return
	}
	t := attr[0].(*TimeTask)
	left := attr[1].(string)
	right := attr[2].(string)
	if left == right {
		t.ModifyTime(2 * time.Second)
		t.Start()
	}
}

func count() (interface{}, error) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	return "1", nil
}

func count2() (interface{}, error) {
	fmt.Println("second:", time.Now().Format("2006-01-02 15:04:05"))
	return "1", nil
}

// initialize option and create instance
options := Regist(WithOptionInterval(2 * time.Second),
		WithOptionValue("1"),
		WithOptionLoop(true),
		WithOptionHook(hook1),
		WithOptionHook(hook2))
	n, err := NewTimeTask(options)
	if err != nil {
		return
	}
n.Add("test", count)
n.Start()
time.Sleep(8 * time.Second)

// create another option and update
options2 := Regist(WithOptionInterval(2 * time.Second),
		WithOptionValue("1"),
		WithOptionLoop(false),
		WithOptionHook(hook3))
n.Update("test", count2, options2)
time.Sleep(15 * time.Second)
```