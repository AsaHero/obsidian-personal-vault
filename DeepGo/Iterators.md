Помнишь классический **off by one error** — самую нелепую и при этом самую живучую ошибку в коде? Она почти всегда вылезает именно в ручных циклах: `for node != nil { ...; node = node.Next }`, `for i := 0; i <= len(s); i++` и подобное. Итераторы, появившиеся в Go начиная с **1.23**, — это как раз способ один раз аккуратно написать такой цикл внутри переиспользуемой функции, а не гонять руками указатели/индексы в каждом месте, где нужен обход.

## Что такое итератор

**Итератор** — это функция, которая передаёт элементы последовательности по одному в функцию обратного вызова, обычно называемую `yield`. Итератор останавливается либо когда последовательность закончилась, либо когда `yield` вернул `false` — то есть потребитель попросил прерваться досрочно (условно, сделал `break`).

Пакет `iter` определяет два сокращения для такого рода функций:

```go
type (
	Seq[V any]     func(yield func(V) bool)
	Seq2[K, V any] func(yield func(K, V) bool)
)
```

`Seq` — итератор с одним значением на элемент, `Seq2` — с парой (обычно ключ/значение или индекс/значение). Произносится `Seq` как "seek" — от "sequence". Синтаксис `[V any]` в объявлении — обычный [[Generics|generic]] type parameter, ничего специфичного для итераторов тут нет.

## range-over-func: самый простой пример

```go
package main

import "fmt"

func Range(yield func(int) bool) {
	for i := 0; i < 10; i++ {
		if !yield(i) {
			return
		}
	}
}

func main() {
	for value := range Range { // да, range можно навесить прямо на функцию!
		fmt.Println(value)
	}
}
```

Как это работает:

- `Range` принимает функцию `yield` и вызывает её на каждой итерации.
- Если `yield` вернул `false` (например, потому что в теле `for range` сработал `break`) — итерация обязана прерваться через `return`.
- В `main` `for value := range Range` под капотом дёргает `Range`, подсовывая ему скрытую `yield`, которая присваивает значение в `value` и выполняет тело цикла.

Вывод — числа от 0 до 9.

## Параметризуем через замыкание

Захардкоженные 10 итераций — не очень полезно. Сделаем `Range` фабрикой, которая настраивает диапазон через замыкание:

```go
package main

import "fmt"

func Range(size int) func(func(int) bool) {
	return func(yield func(int) bool) {
		for i := 0; i < size; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func main() {
	for value := range Range(5) { // 0, 1, 2, 3, 4
		fmt.Println(value)
	}
}
```

Теперь `Range(5)` возвращает готовый итератор `func(yield func(int) bool)`, а не сама является итератором.

### ⚠️ Что будет, если не проверять результат `yield`

Возврат `yield` — не формальность, а часть контракта. Если его игнорировать и продолжать звать `yield` как ни в чём не бывало:

```go
func Bad(yield func(int) bool) {
	for i := 0; i < 5; i++ {
		yield(i) // не проверяем bool!
		// если потребитель уже сделал break на i=0 —
		// вызов yield(1) здесь нарушает контракт итератора
	}
}
```

...то в лучшем случае вы впустую продолжите генерировать значения, которые уже никому не нужны. А в худшем — рантайм детектирует, что `yield` вызван ещё раз **после того**, как он уже однажды вернул `false`, и уронит программу с паникой вида:

```
runtime error: range function continued iteration after function for loop body returned false
```

Правило простое: как только `yield` вернул `false` — сразу `return`, и никогда больше `yield` из этого итератора не звать.

## `iter.Seq`/`iter.Seq2` — это просто типы-хелперы

Возвращаясь к типам из начала: `iter.Seq[V]` и `iter.Seq2[K, V]` — это не какая-то особая магия, а просто именованные типы для сигнатуры, которую иначе пришлось бы каждый раз писать руками:

```go
package main

import (
	"fmt"
	"iter"
	"strconv"
)

func Range1(size int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < size; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func Range2(size int) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i := 0; i < size; i++ {
			if !yield(i, "val: "+strconv.Itoa(i)) {
				return
			}
		}
	}
}

func main() {
	for value := range Range1(5) {
		fmt.Println(value)
	}

	for firstValue, secondValue := range Range2(5) {
		fmt.Println(firstValue, secondValue)
	}

	for firstValue := range Range2(5) { // забираем только первое значение пары
		fmt.Println(firstValue)
	}
}
```

Последний цикл показывает интересную деталь: `Seq2` можно использовать и в однозначном `for range` — тогда Go отдаёт только первое значение пары (условный "ключ"), второе просто отбрасывается.

## Реальный пример: обход слайса в обратном порядке

(вспомни [[Array and Slices#Iteration tricks|итерацию по слайсам]] из конспекта Array and Slices — тут та же идея обхода, только один раз аккуратно завёрнутая в yield)

```go
package main

import (
	"fmt"
	"iter"
)

func Backward[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(s[i]) {
				return
			}
		}
	}
}

func main() {
	data := []int{1, 2, 3, 4, 5}
	for value := range Backward(data) { // slices.Backward делает то же самое
		fmt.Println(value)
	}
}
```

Вывод: `5 4 3 2 1`.

Согласись, код на месте использования стал заметно чище — вместо `for i := len(data) - 1; i >= 0; i--` в каждом месте, где нужен обратный обход, у вас одна переиспользуемая функция. Это ровно тот случай, о котором говорили в начале: если у вас свои графы или деревья, и обход выглядит как `node != nil; node = node.Next` — вынесите этот обход в итератор один раз. Индексная/указательная арифметика, где как раз обычно и рождается off-by-one, пишется и тестируется в одном месте, а не размазывается по всем вызывающим местам.

## Стандартная библиотека уже вовсю на итераторах

### `slices`

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	s := []int{1, 2, 3}

	for i, v := range slices.All(s) { // iter.Seq2[int, T] — индекс + значение
		fmt.Println("item", i, "is", v)
	}

	for v := range slices.Values(s) { // iter.Seq[T] — только значения
		fmt.Println(v)
	}
}
```

Вывод:

```
item 0 is 1
item 1 is 2
item 2 is 3
1
2
3
```

### `maps`

```go
package main

import (
	"fmt"
	"maps"
	"math"
)

func main() {
	m := map[string]float64{
		"pi": math.Pi,
		"e":  math.E,
	}

	for k, v := range maps.All(m) { // iter.Seq2[K, V]
		fmt.Println("constant", k, "is", v)
	}

	for k := range maps.Keys(m) { // iter.Seq[K]
		fmt.Println(k)
	}

	for v := range maps.Values(m) { // iter.Seq[V]
		fmt.Println(v)
	}
}
```

> ⚠️ Порядок вывода при итерации по `map` не гарантирован — при разных запусках `"pi"` и `"e"` могут выводиться в разном порядке. Это давнее и осознанное свойство самого языка (см. [[Maps#Go Allows Edit and Delete While Iteration|поведение map при итерации]]), итераторы над картами тут ничего не меняют.

## Итераторы + горутины: опасный паттерн

(если незнакомо, зачем тут `sync.WaitGroup`/`sync.Mutex` — см. [[Synchronization Primitives|Synchronization Primitives]])

```go
package main

import (
	"fmt"
	"iter"
	"sync"
)

func Range(size int) iter.Seq[int] {
	return func(yield func(int) bool) {
		var wg sync.WaitGroup
		for value := range size { // range over int — тоже фича, из Go 1.22
			wg.Add(1)

			go func() {
				defer wg.Done()
				if !yield(value) { // ❌ yield вызывается параллельно из разных горутин!
					return
				}
			}()
		}

		wg.Wait()
	}
}

func main() {
	for value := range Range(10) {
		fmt.Println(value)
	}
}
```

Этот код содержит сразу две проблемы, и обе — не гипотетические, а реально описанный баг в issue-трекере самого Go (`runtime: confusing panic on parallel calls to yield function`, если захочешь поискать):

1. **Нарушение контракта итератора**: `yield` устроен так, что в любой момент времени может быть активен только **один** его вызов, и вызовы должны идти строго последовательно. Здесь же `yield` дёргается параллельно из десятка независимых горутин — рантайм это детектирует и роняет программу с паникой из той же серии, что и в примере выше (`range function continued iteration...`). Причём паника вылезает "неудобно" и не всегда очевидно откуда, именно поэтому баг и попал в трекер как "confusing panic".
2. **`return` внутри `go func() {...}()` не останавливает внешний цикл**: когда `yield(value)` вернул `false`, `return` завершает только саму анонимную горутину. Остальные уже запущенные горутины как ни в чём не бывало продолжают вызывать `yield`, а `wg.Wait()` терпеливо ждёт их всех — реальной остановки итерации по факту не происходит.

Это не значит, что горутины вообще нельзя использовать внутри итератора — значит, что нельзя **параллелить сам вызов `yield`**.

### Как сделать безопасно

Самый простой (хоть и "в лоб") вариант — синхронизировать так, чтобы `yield` вызывался строго по одному разу за раз:

```go
func Range(size int) iter.Seq[int] {
	return func(yield func(int) bool) {
		var wg sync.WaitGroup
		for value := range size {
			wg.Add(1)

			go func() {
				defer wg.Done()
				if !yield(value) {
					return
				}
			}()

			wg.Wait() // ждём эту горутину, прежде чем стартовать следующую
		}
	}
}
```

Тут кнч честно говоря наивный пример: раз мы ждём каждую горутину сразу после запуска, реального параллелизма как такового уже нет — фактически это тот же последовательный вызов `yield`, просто через горутину. Но зато он безопасен: в любой момент активен только один `yield`.

Более интересный (хоть и всё ещё требующий аккуратности) вариант — дать горутинам реально работать параллельно, но защитить именно момент вызова `yield` мьютексом:

```go
func Range(size int) iter.Seq[int] {
	return func(yield func(int) bool) {
		var wg sync.WaitGroup
		var mu sync.Mutex // защищает только сам вызов yield

		for value := range size {
			wg.Add(1)
			go func(v int) {
				defer wg.Done()
				// ... тут может быть любая тяжёлая параллельная работа ...

				mu.Lock()
				defer mu.Unlock()
				yield(v) // гарантированно не пересечётся с другим yield
			}(value)
		}

		wg.Wait()
	}
}
```

> ⚠️ Даже это не идеальное решение "из коробки": если `yield` вернёт `false` (потребитель сделал `break`), уже запущенные, но ещё не дошедшие до `mu.Lock()` горутины всё равно продолжат работать вхолостую — сигнал "остановись" сюда явно не проброшен. Для честной ранней остановки понадобится что-то вроде `context.Context` или атомарного флага, который каждая горутина проверяет перед вызовом `yield`. В большинстве реальных задач с параллельной обработкой проще собрать результаты в канал и итерировать по нему уже последовательно, чем городить параллельный `yield`.

## Работа с базой данных через итератор

Хочется прятать вызовы `rows.Next()`/`rows.Scan()` за итератором, чтобы не повторять их в каждом месте, где идёт обход результата запроса:

```go
package main

import (
	"database/sql"
	"iter"
)

func doQuery(query string) (*sql.Rows, error) { /* ... */ return nil, nil }

func QueryDB[T any](query string) iter.Seq[T] {
	return func(yield func(T) bool) {
		rows, err := doQuery(query)
		if err != nil {
			return // а что делать с ошибкой? см. ниже
		}
		defer rows.Close()

		for rows.Next() {
			var value T
			if err := rows.Scan(&value); err != nil {
				return // и тут та же проблема
			}
			if !yield(value) {
				return
			}
		}
	}
}
```

Проблема на поверхности: `iter.Seq[T]` умеет отдавать только `T`, и ошибке из `doQuery`/`Scan` просто некуда деться — она тихо теряется.

### Пробуем протащить ошибку через `Seq2`

```go
package main

import (
	"database/sql"
	"iter"
)

type User struct {
	// поля пользователя
}

func doQuery(query string) (*sql.Rows, error) { /* ... */ return nil, nil }

func DoQuery[T any](query string) (iter.Seq2[T, error], error) {
	rows, err := doQuery(query)
	if err != nil {
		return nil, err
	}

	return func(yield func(T, error) bool) {
		defer rows.Close()

		for rows.Next() {
			var value T
			err := rows.Scan(&value)
			if !yield(value, err) {
				return
			}
		}
	}, nil
}

func main() {
	rows, err := DoQuery[User]("SELECT * FROM users")
	if err != nil {
		// handling...
	}

	for user, err := range rows {
		_ = user
		_ = err
	}
}
```

`iter.Seq2[T, error]` — вполне признанный в community паттерн именно для такого случая: ошибка есть не всегда (не на каждой итерации), но потребитель обязан быть готов её увидеть в любой момент, `for value, err := range ...` естественно ложится в существующий синтаксис `for range`.

### ⚠️ Подводный камень: паника до начала итерации = ресурс никогда не закроется

```go
func main() {
	rows, err := DoQuery[User]("SELECT * FROM users")
	if err != nil {
		// handling...
	}

	panic(100) // где-то между получением итератора и началом range

	for user, err := range rows {
		_ = user
		_ = err
	}
}
```

Тут `rows.Close()` никогда не вызовется. Причина не в `panic` как таковой, а в том, **где именно живёт `defer rows.Close()`**: он находится внутри тела замыкания, которое мы вернули как `iter.Seq2[T, error]`. Это тело физически начинает выполняться только тогда, когда стартует `for ... range rows` — то есть `defer` ещё даже не зарегистрирован в момент паники. Классический `defer` сразу после получения ресурса (как мы привыкли с `rows, err := db.Query(...); defer rows.Close()`) тут не работает, потому что сам ресурс (`*sql.Rows`) добывается **лениво**, внутри итератора, а не сразу в момент вызова `DoQuery`.

### Решение: возвращать функцию очистки отдельно от итератора

```go
package main

import (
	"database/sql"
	"iter"
)

type User struct{}

func doQuery(query string) (*sql.Rows, error) { /* ... */ return nil, nil }

func QueryDB[T any](query string) (iter.Seq2[T, error], func(), error) {
	rows, err := doQuery(query)
	if err != nil {
		return nil, nil, err
	}

	return func(yield func(T, error) bool) {
			for rows.Next() {
				var value T
				err := rows.Scan(&value)
				if !yield(value, err) {
					return
				}
			}
		}, func() {
			_ = rows.Close()
		}, nil
}

func main() {
	rows, clean, err := QueryDB[User]("SELECT * FROM users")
	if err != nil {
		// handling...
	}

	defer clean() // теперь defer стоит прямо там же, где ресурс реально был получен

	// panic("here") — теперь rows.Close() всё равно отработает

	for user, err := range rows {
		_ = user
		_ = err
	}
}
```

Тут `clean` — обычная функция, возвращённая сразу вместе с итератором, в тот же момент, когда `*sql.Rows` реально был создан. `defer clean()` ставится сразу на вызывающей стороне — ровно там, где привычно ожидать `defer` для только что открытого ресурса. Теперь неважно, случится паника до начала итерации, во время неё или после — `rows.Close()` гарантированно отработает.

Согласись, уже не так красиво в плане "просто одна функция вместо трёх штук на выходе", но зато честно и предсказуемо. И да, паника такое явление, которое происходит не всегда — но именно поэтому её отсутствие "в большинстве случаев" не оправдывает утечку ресурса в оставшихся.

### Альтернатива: `Err()`-метод вместо `Seq2[T, error]`

Есть и другой, более старый по духу подход — как у `bufio.Scanner`: сам итератор отдаёт только значения, а ошибку можно проверить отдельным методом после цикла.

|.|`iter.Seq2[T, error]`|Отдельный `Err()` (как `bufio.Scanner`)|
|---|---|---|
|Проверка ошибки|На каждой итерации, `for v, err := range ...`|Один раз, после `for v := range ...`|
|Легко забыть проверить|Да — компилятор не заставит проверять `err` каждый раз|Да — можно забыть вызвать `Err()` в конце|
|Синтаксис вызывающей стороны|Чуть многословнее на каждой итерации|Чище внутри цикла, доп. строка в конце|
|Подходит, когда|Ошибка может быть у отдельного элемента (например, парсинг конкретной строки)|Ошибка одна на весь обход целиком (например, разрыв соединения с БД)|

Оба варианта одинаково легитимны, выбор — вопрос вкуса и природы самой ошибки.

## Композиция итераторов — конвейеры в функциональном стиле

А вот тут начинается настоящее веселье:

```go
package main

import (
	"fmt"
	"iter"
)

func Range(size int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for value := range size {
			if !yield(value) {
				return
			}
		}
	}
}

func Mul(seq iter.Seq[int], multiplier int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for value := range seq { // "range" по чужому итератору внутри своего
			if !yield(value * multiplier) {
				return
			}
		}
	}
}

func Filter(seq iter.Seq[int], predicate func(int) bool) iter.Seq[int] {
	return func(yield func(int) bool) {
		for value := range seq {
			if predicate(value) {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func main() {
	data := Range(1_000)
	data = Mul(data, 2)
	data = Filter(data, func(i int) bool { return i%2 == 0 })

	for value := range data {
		fmt.Println(value)
	}
}
```

- `Range(size)` — источник, генерирует `0..size-1`.
- `Mul(seq, multiplier)` — оборачивает `seq`, умножая каждое пролетающее через него значение.
- `Filter(seq, predicate)` — оборачивает `seq`, пропуская наружу только значения, прошедшие `predicate`.

По сути это `map`/`filter` из функциональных языков, но без единой промежуточной аллокации. **Вся эта конструкция ленивая**: ни один элемент реально не вычисляется, пока не начнётся `for value := range data`. Каждое число проходит по цепочке `Range → Mul → Filter → yield` целиком, одно за другим — никаких промежуточных слайсов на 1000 элементов на каждом шаге конвейера, память расходуется так же экономно, как в стриминговых пайплайнах на каналах, но без накладных расходов на сами каналы и переключение горутин.

## Почему переменная называется именно `yield`

Потому что под капотом `range-over-func` реализован через внутренний рантайм-механизм, похожий на **stackless coroutines** — во время итерации выполнение как бы прыгает туда-сюда между телом итератора и телом цикла, вместо того чтобы честно параллелить их в отдельных горутинах. Слово `yield` — устоявшееся имя именно для этой концепции в языках, где есть stackless-корутины/генераторы (Python-генераторы, C# `yield return`, и так далее), поэтому Go community и переняло его для этой же роли.

**Ждём ли мы полноценные corountines в самом Go?** На данный момент (актуально по состоянию на середину 2026 года) в языке нет отдельной публичной языковой конструкции "coroutine" — то, что используется под капотом `range-over-func` (внутренний пакет `runtime`/coro), не выставлено наружу как самостоятельный примитив для разработчиков. Тема периодически всплывает в community-обсуждениях (в том числе есть исследовательские наработки от самой команды Go на этот счёт), но как готовая фича в релизах она пока не появилась — так что пока для explicit-конкурентности у нас всё ещё goroutines + channels + context, а `iter`/`yield` — это именно про удобный обход последовательностей, а не про замену горутин.