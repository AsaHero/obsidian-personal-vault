## Array

### Iteration tricks

Range loop работает с коипей  оригинального массива. Поэтому нужно это учитывать, если внутри Range loop хочешь менять огинальный массив.

### Slice to Array

Можно получить, но данные скопируются в новый участок памяти.

```Go
slice := []int{1, 2, 3, 4, 5}
array := ([4]int)(slice[0:4])
```

Кнч это можно обойти с помощью указателей:
```Go
slice := []int{1, 2, 3, 4, 5}
array := (*[4]int)(slice[0:4])
```

Интересный момент с указателями на массив:

```Go
var arr *[4]int // this is nil for sure

fmt.Println(len(arr)) // this is 4
fmt.Println(cap(arr)) // YES! array also has cap, and here it is 4

// this will work
for idx, _ := range arr {
	fmt.Println(idx)
}

// This will panic in runtime 
for idx, val := range arr {
	fmt.Println(val)
}

```

### Array always in stack ? 
Воперых, да массив создается в стэке, если не учитывать escape analysis (см. также [[Functions#Inlining|оптимизации компилятора]]), только есть ограчниние по памяти и это <=10 МБ.  Если массив больше 10 МБ, то он уходит в HEAP.

### Array in stack has static memory address ?
Нет, может пригыть с одного адресса в другой. Но если вы HEAP, будет статичный. 
## Slices

### Slice under the hood
```Go
type slice struct {
	pointer unsafe.Pointer
	len int
	cap int 
}
```

### Iteration tricks 
Range loop работает с коипей оригинальной структуры слайса, но обе структуры будут указывать на один и тот же массив в памяти.

```Go
x := []string{"A", "M", "C"}

for i, s := range x {
    print(i, s, ";")
    x[i+1] = "M"
    x = append(x, "Z")
    x[i+1] = "Z"
}
```

Ключевой момент:

```Go
x = append(x, "Z")
```

TТригирит **reallocation**, теперь X будеи указывать на другой участок помяти.
### Slice iteration optimization

В одном лупе могу итерировать по 4 элклмента:
```Go
for i := 0; i < len(arr)/4; i += 4 {
	arr[i]
	arr[i + 1]
	arr[i + 2]
	arr[i + 3]
}
```

Это называется ==Loop Widening== 

### Sub-slice

```Go
// temp = data[<index from> : <index to exceptionally>]

data := []int{1, 2, 3, 4}
temp := data[1;3] // [2, 3]

// temp = data[<index from> :] means till the end
data := []int{1, 2, 3, 4}
temp := data[1:] // [2, 3, 4]

// temp = data[: <index to exceptionally>] means from the begining
data := []int{1, 2, 3, 4}
temp := data[:3] // [1, 2, 3]

// temp = data[:] fully subslice
data := []int{1, 2, 3, 4}
temp := data[:] // [1, 2, 3, 4]

// temp = data[<index from> : <index to exceptionally>: <cap index>] we indicate to what index to be a cap, default till the end
data := []int{1, 2, 3, 4, 5}
temp := data[:1:3] // [1] -> len = 1 cap = 4

```


### Modification of copy slices caution!
При изменении длины/ёмкости одного среза, ищменение никак не отражаются на другом срезе! Так как каждый созданый срез это отдельная структура. Хоть они шарят один тот же массив. 

### What happens, if we get sub-slice from an array ?

```Go
array := [...]int{1, 2, 3, 4, 5}
slice := array[1:3]
```

Просто создается структура слайса, внутри кторого у нас указатель на участок памяти того самого массива

### Length is length anyway

Мы не можем обратиться к индексу больше чем длина среза, хоть cap больше чем length.

```Go
data := make([]int, 3, 6)

fmt.Println(data[4]) // panic
```

Но это я тоже могу обойти хитром образом, через подсрезы:
```Go
data := make([]int, 3, 6)

data = data[0:6]

fmt.Println(data[4]) // will not panic 
```

### GC only cleans slice wholly, no partial cleaning!

То есть если мы создаем подсрез из большего слайса, то в  памяти всегда будет тот самый большой слайс, хоть мы его давно не испольуем:
```Go
func onlyNeedData(data []int) []int {
	data := data[:1]
}

func main() {
	data := []int{1, 2, 4, 5, 6, 7, 8, 9, 0}
	
	data = onlyNeedData(data)
	
	// rest logic using data with only needed data
	// but GC will not clean slice's not needed data!
}
```

Хоть используешь только первые два элемента, остальное будет висеть в памяти все равно, что и есть ==Утечка памяти (Memory Leak)== (аналогичная проблема есть и у [[Strings#Memory Leak with strings|строк]])

Что делать тогла, перекопировать в новый слайс:

```Go
func onlyNeedData(data []int) []int {
	data := append([]int(nil), data[:2]...)
}

func main() {
	data := []int{1, 2, 4, 5, 6, 7, 8, 9, 0}
	
	data = onlyNeedData(data)
	
	// rest logic using data with only needed data
	// but GC will clean slice wholll, cuz no body refer to it now
}
```


### Does array of slice always allocated in heap ?

Вопервых, да обычно слайс аллоцируется в heap, только если его размер больше 64 КБ. Если массив слайса меньше или равно 64КБ то массив создается в стеке. Но не забываем про [[Functions#Inlining|escape analysis]]. 

Но есть момент, когда на слайс с размер меньше 64КБ, сдлеать append перевышающий cap, то новый слайс создается в heap, вне зависимости маленького размера!

## Nuances of arrays and slices

### copy() misleading behavior

Функция copy() копирует один слайс в другой по следующим правилам:

Число элементов, скопированных в другой слайс, определяется минимумом между длиной первого слайса и второго слайса.

То есть этот работает не правильно:
```Go
src := []int{1, 2, 3, 4, 5}
dst := []int{}

copy(dst, src)

fmt.Println(dst) // []
```

Будет правильно след варианты:
```Go
src := []int{1, 2, 3, 4, 5}
dst :=  make(src, len(src))

copy(dst, src)

fmt.Println(dst) // [1, 2, 3, 4, 5]
```

Можно скопировать только 3 элемента:
```Go
src := []int{1, 2, 3, 4, 5}
dst :=  make(src, 3)

copy(dst, src)

fmt.Println(dst) // [1, 2, 3]
```

Также можно использовать append():
```Go
src := []int{1, 2, 3, 4, 5}
dst :=  append([]int(nil), src...)


fmt.Println(dst) // [1, 2, 3, 4, 5]
```

Либо новый метод в стандартной библиотеке **slices.Clone()**:
```Go
src := []int{1, 2, 3, 4, 5}
dst := slices.Clone(src)

// as well as we can do this:
// dst := slices.Clone(src[1:3])

fmt.Println(dst) // [1, 2, 3, 4, 5]
```


### Dangerous Memory Leak

```Go
func main() {
	var number = make([]int, 1<<30) // A LOT OF MEMORY ALLOCATION
	
	pointerToElelm := FindElemnt(numbers, 0) 
}
```

В этом случай у нас всего один указатель который указывает на один эдемент внутри слайсы с размером больше 1ГБ, но GC не очистит память. 