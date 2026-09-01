## WaitGroup

Смотри, есть банальная проблема — запустили [[Goroutines|горутины]], а как узнать что они все отработали?

```Go
package main

import "log"

func main() {
	for i := 0; i < 5; i++ {
		go func() {
			log.Println("test")
		}()
	}
	// тут main может завершиться раньше, чем горутины успеют напечатать
}
```

Можно было бы вставить `time.Sleep`, но это костыль — мы гадаем сколько спать. Для таких задач в Go есть примитивы синхронизации, и `sync.WaitGroup` — самый простой из них.

Работает как счётчик: `Add` увеличивает, `Done` уменьшает, `Wait` блокируется, пока счётчик не станет нулём.

```Go
package main

import (
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("test")
		}()
	}

	wg.Wait()
}
```

Если заранее знаешь сколько горутин запустишь — лучше сразу вызвать `Add(n)` один раз, а не по одному в цикле:

```Go
package main

import (
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			log.Println("test")
		}()
	}

	wg.Wait()
}
```

**Почему это важнее, чем кажется:** внутри `Add` не просто `n++`, а атомарный инкремент, а это на практике примерно в 15 раз медленнее обычного инкремента. Если можешь сразу сказать «я запущу 5 горутин» — скажи один раз, не дёргай `Add` в цикле.

### API

```Go
type WaitGroup struct { ... }

func (wg *WaitGroup) Add(delta int) // увеличивает счетчик
func (wg *WaitGroup) Done()         // уменьшает счетчик на единицу
func (wg *WaitGroup) Wait()         // блокируется, пока счетчик не обнулится
```

### ⚠️ Подводный камень: копирование WaitGroup

```Go
package main

import "sync"

func done(wg sync.WaitGroup) { // ❌ передали по значению!
	wg.Done()
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	done(wg)
	wg.Wait() // повиснет навсегда
}
```

Тут `wg` передаётся **по значению**, то есть внутри `done` мы уменьшаем счётчик у _копии_, а оригинал в `main` как ждал единицу, так и ждёт — deadlock.

**Правило на всю жизнь:** никогда не копируем ничего из пакета `sync`. Всегда передаём указатель:

```Go
package main

import "sync"

func done(wg *sync.WaitGroup) { // ✅ указатель
	wg.Done()
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	done(&wg)
	wg.Wait()
}
```

---

## Mutex

### Сначала — классика с race condition

```Go
package main

import (
	"fmt"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1000)

	value := 0
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			value++
		}()
	}

	wg.Wait()

	fmt.Println(value) // почти никогда не будет 1000
}
```

**Почему так происходит.** `value++` это не одна атомарная операция, а три:

1. read — загрузить значение из памяти в регистр
2. add — прибавить 1
3. write — записать новое значение обратно в память

И между этими тремя шагами может «пройти» сколько угодно других горутин (условно — «тысяча лет и тысяча горутин» между любыми двумя строчками кода, если нет синхронизации).

Классическая аномалия — **потеря обновления (lost update)**:

```
value = 100

G1:  oldValue := value        // oldValue = 100
G2:  oldValue := value        // oldValue = 100

G1:  newValue := oldValue + 1 // newValue = 101
     value = newValue         // value = 101

G2:  newValue := oldValue + 1 // newValue = 101
     value = newValue         // value = 101

// итог: value = 101, хотя инкрементов было два
```

### Мьютекс

**Mutual Exclusion** — взаимное исключение. У нас есть критическая секция кода, и мы хотим, чтобы в ней в один момент времени работала только одна горутина, а остальные ждали, пока она не скажет «я вышла».

```Go
type Mutex struct { ... }

func (m *Mutex) Lock()          // захватить мьютекс
func (m *Mutex) Unlock()        // освободить мьютекс
func (m *Mutex) TryLock() bool  // попробовать захватить мьютекс не блокируясь
```

Правильный инкремент:

```Go
package main

import (
	"fmt"
	"sync"
)

func main() {
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(1000)

	value := 0
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()

			mutex.Lock()
			value++
			mutex.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println(value) // теперь стабильно 1000
}
```

Ещё пример — мьютекс защищает не переменную саму по себе, а _участок кода_, который к ней обращается:

```Go
package main

import (
	"fmt"
	"sync"
)

var mutex sync.Mutex
var value string

func set(v string) {
	mutex.Lock()
	value = v // пока G1 тут — G2 не может ни читать, ни писать
	mutex.Unlock()
}

func print() {
	mutex.Lock() // G2 подождёт, потому что мьютекс общий (глобальный)
	fmt.Println(value)
	mutex.Unlock()
}
```

### ⚠️ А это вообще корректно?

```Go
var flag bool

func goroutine1() {
	flag = true
}

func goroutine2() {
	flag = false
}
```

Казалось бы, `bool` — это один байт, чтение/запись атомарны почти на любом железе, что тут может пойти не так? Проблема в том, что:

- атомарность одной операции чтения/записи на одной архитектуре **не гарантирует** того же на другой;
- компилятор и процессор имеют право **переупорядочивать инструкции** ради оптимизации — они не знают, что это критическая секция.

**Best practice:** старайся вообще не допускать data race в коде, даже там, где «вроде бы не страшно». Такой баг может годами не стрелять, а потом внезапно выстрелит на выросшем в 10 раз кодовой базе — искать его будешь месяцами. Это и называют «термоядерным багом».

---

## Data Race vs Race Condition — не путать!

**Data Race** — несинхронизированное обращение к одному участку памяти из разных горутин, при этом хотя бы одна из них выполняет запись.

- Если все горутины только читают → data race физически невозможен.
- Как только появляется хотя бы одна запись без синхронизации → появляется риск data race.

**Race Condition** — это уже ошибка _проектирования_: результат работы системы зависит от того, в каком порядке выполнились части кода. Причём race condition может быть даже без единого data race!

Пример: тут нет data race (мьютекс защищает и запись, и чтение), но race condition — есть:

```Go
package main

import (
	"fmt"
	"sync"
)

func main() {
	text := ""
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		mx.Lock()
		text = "hello world"
		mx.Unlock()
	}()

	go func() {
		defer wg.Done()
		mx.Lock()
		fmt.Println(text)
		mx.Unlock()
	}()

	wg.Wait()
}
```

При запуске то увидим пустую строку, то `hello world`. Мьютекс тут ни при чём — он честно защищает доступ к памяти (data race устранён), но **порядок выполнения двух горутин не определён**, поэтому какая из них первая захватит мьютекс — вопрос удачи. Похожая гонка, но вокруг обычной переменной, а не мьютекса — разобрана в [[Channels#Data race, который не про сам канал|Channels]].

> 💡 Мьютекс лечит data race, но не лечит race condition, если у тебя логика зависит от порядка запуска горутин. Тут нужна отдельная синхронизация порядка (например, через канал или WaitGroup по этапам).

### Практика: инкапсулируй мьютекс внутри типа

Хорошая идея — не таскать мьютекс по всему коду, а спрятать его внутри структуры вместе с данными, которые он защищает.

```Go
package main

import "sync"

type Cache struct {
	mutex sync.Mutex
	data  map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
}

func (c *Cache) Get(key string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Size() > 0 { // ⚠️ deadlock!
		return c.data[key]
	}

	return ""
}

func (c *Cache) Size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return len(c.data)
}
```

### ⚠️ Подводный камень — рекурсивная блокировка

Смотри внимательно на `Get`: мы уже захватили `c.mutex.Lock()`, а внутри дёргаем `c.Size()`, который **тоже** пытается захватить тот же мьютекс. `sync.Mutex` в Go **не реентерабельный** (не поддерживает повторный захват той же горутиной) — получаем гарантированный deadlock. Мьютекс сам по себе не знает, кто именно его держит, поэтому даже «та же горутина» не может зайти второй раз.

Решение — не звать публичный метод с блокировкой изнутри другого метода с блокировкой, а вынести приватную версию без Lock:

```Go
func (c *Cache) size() int { // без блокировки, для внутреннего использования
	return len(c.data)
}
```

---

## Deadlock

Ситуация, когда несколько горутин ждут ресурсы, занятые друг другом, и никто не может продолжить выполнение.

```Go
package main

import "sync"

var resource1 int
var resource2 int

func normalizeResources(lhs, rhs *sync.Mutex) {
	lhs.Lock()
	rhs.Lock()

	// normalization

	rhs.Unlock()
	lhs.Unlock()
}

func main() {
	var mutex1 sync.Mutex
	var mutex2 sync.Mutex

	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			normalizeResources(&mutex1, &mutex2)
		}()
	}

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			normalizeResources(&mutex2, &mutex1) // ⚠️ порядок захвата обратный!
		}()
	}

	wg.Wait()
}
```

**Почему стреляет не сразу.** Этот код может месяцами работать нормально, а потом внезапно повиснуть, потому что первая группа горутин захватывает `mutex1 → mutex2`, а вторая — `mutex2 → mutex1`. Если в какой-то момент первая группа захватит `mutex1`, а вторая в этот же момент захватит `mutex2` — обе зависнут в ожидании друг друга навсегда.

> ✅ **Правило:** всегда захватывай мьютексы в одном и том же порядке во всём приложении.

### ⚠️ Deadlock не всегда паникует рантаймом

Многие думают: «если будет deadlock, Go рантайм сам это обнаружит и упадёт с паникой». Это не всегда так — паника `fatal error: all goroutines are asleep - deadlock!` возникает только когда **абсолютно все** горутины в проекте заснули. Если хотя бы одна горутина продолжает что-то делать (например, тикает по таймеру) — рантайм не считает это deadlock'ом, хотя часть горутин реально повисла навечно:

```Go
package main

import (
	"log"
	"sync"
	"time"
)

var resource1 int
var resource2 int

func normalizeResources(lhs, rhs *sync.Mutex) {
	lhs.Lock()
	rhs.Lock()
	rhs.Unlock()
	lhs.Unlock()
}

func main() {
	var mutex1 sync.Mutex
	var mutex2 sync.Mutex

	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			normalizeResources(&mutex1, &mutex2)
		}()
	}

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			normalizeResources(&mutex2, &mutex1)
		}()
	}

	time.Sleep(time.Millisecond * 100)

	go func() {
		for {
			time.Sleep(time.Second)
			log.Println("tick") // эта горутина жива — рантайм не запаникует
		}
	}()

	wg.Wait()
}
```

Вывод: не полагайся на то, что рантайм всегда подскажет о deadlock'е. Нужны отдельные инструменты (race detector, профилирование горутин, мониторинг) чтобы это ловить.

---

## Livelock

Система не «застревает», а **занимается бесполезной работой**: состояние постоянно меняется, но полезной работы не происходит.

```Go
package main

import (
	"fmt"
	"runtime"
	"sync"
)

var mutex1 sync.Mutex
var mutex2 sync.Mutex

func goroutine1() {
	mutex1.Lock()

	runtime.Gosched()
	for !mutex2.TryLock() {
		// активное ожидание — "вежливо" уступаем, но всё равно крутимся
	}

	mutex2.Unlock()
	mutex1.Unlock()

	fmt.Println("goroutine1 finished")
}

func goroutine2() {
	mutex2.Lock()

	runtime.Gosched()
	for !mutex1.TryLock() {
		// активное ожидание
	}

	mutex1.Unlock()
	mutex2.Unlock()

	fmt.Println("goroutine2 finished")
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		goroutine1()
	}()

	go func() {
		defer wg.Done()
		goroutine2()
	}()

	wg.Wait()
}
```

Тут обе горутины пытаются «вежливо» уступить друг другу через `TryLock` в цикле, но при неудачном стечении обстоятельств могут бесконечно долго крутиться, постоянно захватывая/освобождая, но так и не завершая работу. В отличие от deadlock, тут CPU активно грузится — просто без толку.

---

## Starvation

Горутина не может получить нужные ей ресурсы для работы («голодает»), потому что их постоянно забирают более «жадные» горутины.

> 💡 Starvation бывает не только с мьютексами — процессор, память, файловые дескрипторы, коннекты к базе данных — всё это тоже ресурсы, за которые можно «голодать».

```Go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	const runtime = 1 * time.Second

	greedyWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			mutex.Lock()
			time.Sleep(3 * time.Nanosecond)
			mutex.Unlock()
			count++
		}

		fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
	}

	politeWorker := func() {
		defer wg.Done()

		var count int
		for begin := time.Now(); time.Since(begin) <= runtime; {
			mutex.Lock()
			time.Sleep(1 * time.Nanosecond)
			mutex.Unlock()

			mutex.Lock()
			time.Sleep(1 * time.Nanosecond)
			mutex.Unlock()

			mutex.Lock()
			time.Sleep(1 * time.Nanosecond)
			mutex.Unlock()

			count++
		}

		fmt.Printf("Polite worker was able to execute %v work loops.\n", count)
	}

	wg.Add(2)
	go greedyWorker()
	go politeWorker()

	wg.Wait()
}
```

`greedyWorker` держит мьютекс дольше за раз (один длинный `Lock`), а `politeWorker` берёт и отпускает мьютекс три раза короткими порциями. В итоге `politeWorker` чаще проигрывает гонку за мьютекс — жадный воркер успевает выполнить больше циклов. Мораль: держи критическую секцию максимально короткой.

---

## Cond (условная переменная)

Примитив синхронизации, который блокирует одну или несколько горутин до того момента, пока другая горутина не просигналит о выполнении какого-то условия.

### API

```Go
type Cond struct {
	L Locker
}

func NewCond(l Locker) *Cond // создает Cond с использованием мьютекса
func (c *Cond) Broadcast()   // оповещает все горутины
func (c *Cond) Signal()      // оповещает одну горутину
func (c *Cond) Wait()        // ожидает наступления сигнала
```

```Go
package main

import (
	"log"
	"sync"
	"time"
)

func subscribe(name string, data map[string]string, c *sync.Cond) {
	c.L.Lock()

	for len(data) == 0 {
		c.Wait()
	}

	log.Printf("[%s] %s\n", name, data["key"])

	c.L.Unlock()
}

func publish(name string, data map[string]string, c *sync.Cond) {
	time.Sleep(time.Second)

	c.L.Lock()
	data["key"] = "value"
	c.L.Unlock()

	log.Printf("[%s] data publisher\n", name)
	c.Broadcast()
}

func main() {
	data := map[string]string{}
	cond := sync.NewCond(&sync.Mutex{})

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		subscribe("subscriber_1", data, cond)
	}()

	go func() {
		defer wg.Done()
		subscribe("subscriber_2", data, cond)
	}()

	go func() {
		defer wg.Done()
		publish("publisher", data, cond)
	}()

	wg.Wait()
}
```

### ⚠️ Как НЕ надо пользоваться Cond

```Go
package main

import "sync"

func waitWithoutLock() {
	cond := sync.NewCond(&sync.Mutex{})
	cond.Wait() // ❌ паника: Wait без Lock
}

func waitAfterSignal() {
	cond := sync.NewCond(&sync.Mutex{})
	cond.Signal() // сигнал ушёл "в никуда" — слушателей ещё нет

	cond.L.Lock()
	cond.Wait() // а тут уже никто не разбудит
	cond.L.Unlock()
}

func main() {}
```

`Wait()` обязательно вызывается **под захваченным мьютексом**. А `Signal()`/`Broadcast()`, отправленные до того, как кто-то встал в `Wait()`, просто теряются — Cond не запоминает «прошлые» сигналы.

### Что реально происходит внутри Wait()

```Go
c.checker.check()
t := runtime_notifyListAdd(&c.notify)
c.L.Unlock()
runtime_notifyListWait(&c.notify, t)
c.L.Lock()
```

Обрати внимание: `Wait()` сам временно отпускает мьютекс на время ожидания и захватывает обратно перед возвратом. Именно поэтому стандартный паттерн — **всегда вызывать Wait внутри цикла**, а не в `if`:

```Go
c.L.Lock()
for !condition() {
	c.Wait()
}
... используем condition ...
c.L.Unlock()
```

**Почему цикл, а не `if`:** пока горутина спала, условие могло снова стать ложным (например, другая горутина уже забрала ресурс), плюс возможны «ложные пробуждения» — поэтому после каждого `Wait()` условие нужно перепроверять.

### Семафор на основе Cond

Примитив, который ограничивает число горутин, одновременно работающих с общим ресурсом.

```Go
package main

import (
	"sync"
)

type Semaphore struct {
	count     int
	max       int
	condition *sync.Cond
}

func NewSemaphore(limit int) *Semaphore {
	mutex := &sync.Mutex{}
	return &Semaphore{
		max:       limit,
		condition: sync.NewCond(mutex),
	}
}

func (s *Semaphore) Acquire() {
	s.condition.L.Lock()
	defer s.condition.L.Unlock()

	for s.count >= s.max {
		s.condition.Wait()
	}

	s.count++
}

func (s *Semaphore) Release() {
	s.condition.L.Lock()
	defer s.condition.L.Unlock()

	s.count--
	s.condition.Signal()
}
```

---

## Атомарные операции

Операция, которая либо выполняется целиком, либо не выполняется вовсе — без «промежуточных» состояний, видимых другим горутинам.

Есть два стиля работы с пакетом `sync/atomic`:

- **процедурный (функциональный)** — старый стиль, функции вида `atomic.AddInt32(&x, 1)`
- **ООП-стиль** — новый, через типы вроде `atomic.Int32`, `atomic.Bool` и методы на них

### Процедурный стиль

Второй способ решить нашу изначальную проблему гонки на инкременте:

```Go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1000)

	var value int32
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			atomic.AddInt32(&value, 1)
		}()
	}

	wg.Wait()

	fmt.Println(value)
}
```

Важный момент: мы передаём **адрес** участка памяти, и именно на этом адресе процессор выполняет атомарную операцию (специальной CPU-инструкцией, а не через мьютекс).

### API (процедурный стиль)

```Go
func AddInt32(addr *int32, delta int32) (new int32)         // добавить значение
func LoadInt32(addr *int32) int32                            // получить значение
func StoreInt32(addr *int32, val int32)                      // установить значение
func SwapInt32(addr *int32, new int32) (old int32)           // заменить значение и вернуть старое
func CompareAndSwapInt32(addr *int32, old, new int32) bool   // заменить на new, если текущее значение == old
```

### ООП-стиль

```Go
type Bool struct { ... }

func (x *Bool) CompareAndSwap(old, new bool) (swapped bool)
func (x *Bool) Load() bool
func (x *Bool) Store(val bool)
func (x *Bool) Swap(new bool) (old bool)
```

### Зачем нужен Compare-And-Swap (CAS)

Казалось бы, можно было решить задачу «инициализировать один раз» так:

```Go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var data map[string]string
var initialized atomic.Bool

func initialize() {
	if !initialized.Load() { // ⚠️ TOCTOU: между Load и Store может влезть другая горутина
		initialized.Store(true)
		data = make(map[string]string)
		fmt.Println("initialized")
	}
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			initialize()
		}()
	}

	wg.Wait()
}
```

На первый взгляд вроде рабочий код, и в 99% случаев он и правда отработает как надо. Но это классический race condition: между `Load()` и `Store()` может пройти переключение контекста — Go scheduler в курсе про свои горутины, но есть ещё **OS scheduler**, который может вытеснить поток ОС ровно между этими двумя строками. В итоге пара горутин одновременно увидят `false` и обе начнут инициализацию.

Решение — атомарный **Compare-And-Swap**: сравнить и заменить одной неделимой операцией. Практический пример именно такого CAS-паттерна для потокобезопасной отмены — самописная реализация `cancel()` в [[Context#Своя реализация Context (для понимания, как это устроено внутри)|Context]].

```Go
func initialize() {
	if initialized.CompareAndSwap(false, true) { // проверка и запись — одной атомарной операцией
		data = make(map[string]string)
		fmt.Println("initialized")
	}
}
```

Теперь только одна горутина сможет успешно "перевести" `false → true`, все остальные получат `false` от CAS и просто пропустят инициализацию.

### CAS Loop

Типичный паттерн — крутить `Load → вычислить новое значение → CompareAndSwap` в цикле, пока не получится:

```Go
package main

import (
	"sync/atomic"
)

func IncrementAndGet(pointer *int32) int32 {
	for {
		currentValue := atomic.LoadInt32(pointer)
		nextValue := currentValue + 1
		if atomic.CompareAndSwapInt32(pointer, currentValue, nextValue) {
			return nextValue
		}
		// если CAS не удался — кто-то нас опередил, пробуем ещё раз
	}
}
```

Цикл может прокрутиться не один раз, а несколько — это нормально при высокой конкуренции. На таких CAS-циклах построено большинство **lock-free структур данных**.

### ⚠️ Подводный камень CAS-циклов: та же проблема «1000 лет между строк»

Узнав про CAS, многие начинают лепить его куда попало и забывают главное правило конкурентности: между любыми двумя строчками кода может пройти сколько угодно других горутин.

```Go
package main

import "sync/atomic"

type Data struct {
	count atomic.Int32
}

func (d *Data) Process() {
	d.count.Add(1) // допустим тут count стал 100
	// ⚠️ ровно тут может случиться context switch, и другая горутина
	// успеет ещё раз сделать Add и, например, сбросить count в 0
	if d.count.CompareAndSwap(100, 0) {
		// do something...
	}
}
```

Проблема в том, что `Add` и последующий `CompareAndSwap` — это **две отдельные** атомарные операции, а между ними значение может измениться сколько угодно раз.

**Как чинить:** используем возвращаемое значение самого `Add`, а не перечитываем его отдельным вызовом — так вся логика опирается на одну атомарную операцию:

```Go
package main

import "sync/atomic"

type Data struct {
	count atomic.Int32
}

func (d *Data) Process() {
	v := d.count.Add(1) // v — это именно то значение, которое получилось ИМЕННО из нашего Add
	if v == 100 {
		d.count.Add(-100)
	}
}
```

---

## False Sharing

Тонкая и очень болезненная штука, если гонишься за перфомансом на многоядерных системах.

### Бенчмарк-заготовка: Mutex vs Atomic vs Sharded Atomic

```Go
package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// go test -bench=. perf_test.go

type MutexCounter struct {
	value int32
	mutex sync.Mutex
}

func (c *MutexCounter) Increment(int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value++
}

func (c *MutexCounter) Get() int32 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.value
}

type AtomicCounter struct {
	value atomic.Int32
}

func (c *AtomicCounter) Increment(int) {
	c.value.Add(1)
}

func (c *AtomicCounter) Get() int32 {
	return c.value.Load()
}

// шардируем счётчик — у каждой горутины свой "слот"
type ShardedAtomicCounter struct {
	shards [10]AtomicCounter
}

func (c *ShardedAtomicCounter) Increment(idx int) {
	c.shards[idx].value.Add(1)
}

func (c *ShardedAtomicCounter) Get() int32 {
	var value int32
	for idx := 0; idx < 10; idx++ {
		value += c.shards[idx].Get()
	}
	return value
}
```

Идея шардирования: раз каждый воркер бьётся за одну и ту же ячейку памяти (**true sharing**), давай дадим каждому воркеру свою собственную ячейку в массиве — тогда они не должны мешать друг другу.

Результаты бенчмарка (примерные, порядок величин важнее точных цифр):

|Вариант|Время на операцию|
|---|---|
|Mutex|~1708 ns/op|
|Atomic (общая переменная)|~750 ns/op|
|Sharded Atomic (без padding)|~650 ns/op|

Шардирование ускорило, но не так сильно, как хотелось бы — «а лишь бы чуть быстрее». Дальше начинается магия.

### А если добавить padding под размер кэш-линии?

```Go
type AtomicCounter struct {
	value atomic.Int32
	_     [60]byte // padding
}
```

Результат: **14.17 ns/op** — это в разы(!) быстрее.

А если ещё подогнать padding точно под размер кэш-линии CPU:

```Go
type AtomicCounter struct {
	value atomic.Int32
	_     [124]byte // padding
}
```

Результат: **5.6 ns/op**.

### Почему так — разбираемся в магии

**True sharing.** В первой версии (без шардирования) оба CPU реально обращаются к одной и той же ячейке памяти — это честная конкуренция за один и тот же байт/слово в памяти:

```
   ┌──────────┐                    ┌──────────┐
   │   CPU    │                    │   CPU    │
   └────┬─────┘                    └────┬─────┘
        │                               │
        ▼                               ▼
 ┌─────────────────┐            ┌─────────────────┐
 │[■]│ │ │ │ │ │ │ │            │[■]│ │ │ │ │ │ │ │← Cache lines
 └──┬──────────────┘            └──┬──────────────┘
    │                              │
    │        ┌──────────────┐      │
    └───────►│              │◄─────┘
             ▼              ▼
 ┌──────────────────────────────────────────────┐
 │ │ │ │ │ │ │ │ │[■]│ │ │ │ │ │ │ │ │ │ │ │ │ ││← Main Memory
 └──────────────────────────────────────────────┘
                    ▲
          Оба CPU обращаются к ОДНОЙ
          и той же ячейке памяти
```

**False sharing.** Когда мы сделали шардирование _без padding_, каждый шард занимает всего 4 байта (`int32`), а кэш-линия обычно 64 или 128 байт. То есть несколько соседних шардов физически попадают **в одну и ту же кэш-линию**. Когда одно ядро инкрементирует «свой» шард, оно инвалидирует кэш-линию у другого ядра — даже если то ядро с этим конкретным байтом вообще не работало!

```
   ┌──────────┐                    ┌──────────┐
   │   CPU    │                    │   CPU    │
   └────┬─────┘                    └────┬─────┘
        │                               │
        ▼                               ▼
 ┌─────────────────┐            ┌─────────────────┐
 │[■]│ │ │ │ │ │ │ │            │ │[■]│ │ │ │ │ │ │← Cache lines
 └──┬──────────────┘            └─┬───────────────┘
    │                             │
    │      ┌──────────────┐      │
    └─────►│              │◄─────┘
           ▼              ▼
 ┌──────────────────────────────────────────────┐
 │ │ │ │ │ │ │ │ │[■][■]│ │ │ │ │ │ │ │ │ │ │ │ │← Main memory
 └──────────────────────────────────────────────┘
                  ▲   ▲
                  │   │
            CPU1──┘   └──CPU2
         (разные ячейки, но в ОДНОЙ кэш-линии)
```

Это и называется **false sharing** — данные разные, а кэш-линия одна, и ядра всё равно постоянно синхронизируют кэши между собой (протокол когерентности кэша типа MESI), хотя реального пересечения данных нет.

**Padding решает это буквально**: раздвигая структуры так, чтобы каждый шард занимал свою собственную кэш-линию целиком, мы гарантируем, что инкремент одного ядра больше не трогает кэш-линию другого ядра.

> 💡 **Плюсы padding:** огромный прирост перфоманса в highly-contended сценариях на многоядерных машинах. **Минусы:** тратим больше памяти (padding это чистый оверхед), код становится менее очевидным — если не знаешь про false sharing, непонятно зачем там «мёртвые» байты. Плюс padding — это низкоуровневая оптимизация, имеет смысл возиться с ней только когда профилирование реально показало проблему в кэш-линиях, а не «просто на всякий случай».

---

## RWMutex (Shared Mutex)

Примитив, который позволяет **нескольким горутинам одновременно читать** общий ресурс, но **блокирует запись только для одной** горутины за раз (и во время записи блокируются вообще все — и читатели, и писатели).

```
ЧТЕНИЕ (RLock) — доступ разрешён нескольким потокам одновременно:

  Goroutine 1 ──┐
  Goroutine 2 ──┼──► [ Общий ресурс ] ✅ все читают параллельно
  Goroutine 3 ──┘


ЗАПИСЬ (Lock) — доступ только одному потоку, остальные ждут:

  Goroutine 1 ──► [ Общий ресурс ]   🔒 пишет (эксклюзивно)
  Goroutine 2 ──► ⏳ ждёт
  Goroutine 3 ──► ⏳ ждёт
```

```Go
package main

import "sync"

type Counters struct {
	mu sync.RWMutex
	m  map[string]int
}

func (c *Counters) Load(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, found := c.m[key]
	return value, found
}

func (c *Counters) Store(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = value
}
```

### ⚠️ RWMutex — не бесплатный

Логика подсказывает «читаем часто, пишем редко → всегда бери RWMutex», но у RWMutex есть свой оверхед на управление счётчиком читателей:

```Go
package main

import (
	"sync"
	"testing"
)

// go test -bench=. perf_test.go

func BenchmarkMutexAdd(b *testing.B) {
	var number int32
	var mutex sync.Mutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		number++
		mutex.Unlock()
	}
}

func BenchmarkRWMutexAdd(b *testing.B) {
	var number int32
	var mutex sync.RWMutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		number++
		mutex.Unlock()
	}
}
```

|Примитив|Время на операцию|
|---|---|
|`sync.Mutex`|~2.431 ns/op|
|`sync.RWMutex`|~4.774 ns/op|

На чистой записи (без параллельного чтения) `RWMutex` почти в 2 раза медленнее обычного `Mutex`. Профит от `RWMutex` появляется только там, где **чтений реально сильно больше, чем записей**, и они выполняются параллельно. Если не уверен в этом — бери обычный `Mutex`, он проще и дешевле.

---

## Резюме — что когда брать

(смотри также отдельное сравнение [[Channels#Channels vs Mutex — когда что использовать|каналы vs mutex]], если конкретно этот выбор — твой случай)

|Задача|Инструмент|
|---|---|
|Дождаться завершения группы горутин|`sync.WaitGroup`|
|Защитить критическую секцию (чтение + запись вперемешку)|`sync.Mutex`|
|Много читателей, мало писателей|`sync.RWMutex` (но сначала измерь!)|
|Ждать наступления сложного условия|`sync.Cond`|
|Ограничить число одновременных «пользователей» ресурса|Semaphore (на Cond или буферизованном канале)|
|Простой счётчик / флаг без сложной логики|`sync/atomic` (`atomic.Int32`, `atomic.Bool`, ...)|
|Атомарный счётчик на много ядер с высокой конкуренцией|Sharded atomic + padding под cache line|