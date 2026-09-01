**Context** — это набор мета-данных, ассоциированных с конкретным запросом или процессом. На практике он решает две задачи одновременно: сигнализирует об отмене/дедлайне и переносит request-scoped данные (типа `trace_id`) через цепочку вызовов.

## Функции пакета `context`

```go
func Background() Context
func TODO() Context
func WithValue(parent Context, key, val any) Context
func WithoutCancel(parent Context) Context
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithCancelCause(parent Context) (ctx Context, cancel CancelCauseFunc)
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc)
func WithDeadlineCause(parent Context, d time.Time, cause error) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithTimeoutCause(parent Context, timeout time.Duration, cause error) (Context, CancelFunc)
```

Все они, кроме `Background`/`TODO`, принимают parent-context первым аргументом — так строится дерево контекстов (об этом ниже).

## Context — это интерфейс, а не структура

> ⚠️ В исходных заметках `Context` был описан как `struct` с методами на указателе. На самом деле в стандартной библиотеке это **интерфейс**:

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool) // время дедлайна и был ли он вообще установлен
	Done() <-chan struct{}                    // канал, который закроется при отмене/дедлайне
	Err() error                                // причина, по которой канал Done() закрылся
	Value(key any) any                         // значение по ключу из цепочки контекстов
}
```

Именно поэтому можно писать свою реализацию — достаточно реализовать эти 4 метода (сделаем это ниже, для практики).

## `WithTimeout` vs `WithDeadline`

Разница ровно одна — относительное время против абсолютного. Собственно, `WithTimeout` — это просто тонкая обёртка над `WithDeadline`:

```go
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}
```

- `WithTimeout(parent, timeout)` — "отмени через `timeout`" (относительно текущего момента)
- `WithDeadline(parent, d)` — "отмени в конкретный момент времени `d`" (абсолютное время)

|Функция|Когда использовать|Отменяется|
|---|---|---|
|`WithCancel`|нужна ручная отмена по бизнес-событию (юзер закрыл соединение, пришёл сигнал)|только вызовом `cancel()`|
|`WithTimeout`|нужно ограничить операцию по длительности ("не больше 3 секунд")|по истечении таймаута ИЛИ вызовом `cancel()`|
|`WithDeadline`|нужен жёсткий дедлайн на конкретный момент ("не позже 12:00")|по достижении дедлайна ИЛИ вызовом `cancel()`|

## `WithCancel` — пример

```go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func receiveWeather(ctx context.Context, result chan struct{}, idx int) {
	randomTime := time.Duration(rand.Intn(5000)) * time.Millisecond

	timer := time.NewTimer(randomTime)
	defer timer.Stop()

	select {
	case <-timer.C:
		fmt.Printf("finished: %d\n", idx)
		result <- struct{}{}
	case <-ctx.Done(): // как только вызовут cancel(), этот case станет ready
		fmt.Printf("canceled: %d\n", idx)
	}
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(10)

	ctx, cancel := context.WithCancel(context.Background())

	result := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer wg.Done()
			receiveWeather(ctx, result, idx)
		}(i)
	}

	<-result // ждём первый успешный ответ
	cancel() // и отменяем остальные 9 горутин — им незачем дальше работать

	wg.Wait()
}
```

### ⚠️ Отменять контекст нужно там же, где его создал

Вызывать `cancel()` из совершенно другого места кода (например, передав функцию отмены глубоко вниз по стеку в отдельную горутину) считается анти-паттерном — теряется контроль над тем, кто и когда реально управляет жизненным циклом контекста.

Похожая идея касается и того, как **читать** отмену внутри функции:

```go
package main

import "context"

// ❌ Неправильно — игнорируем ctx полностью
func incorrectCheck(ctx context.Context, stream <-chan string) {
	data := <-stream // если stream никогда ничего не пришлёт — зависнем навсегда, ctx тут бесполезен
	_ = data
}

// ✅ Правильно — даём ctx шанс прервать ожидание
func correctCheck(ctx context.Context, stream <-chan string) {
	select {
	case data := <-stream:
		_ = data
	case <-ctx.Done():
		return
	}
}
```

## Как понять, что контекст уже отменён

Как только `Done()` закроется, `ctx.Err()` начинает возвращать не-`nil` ошибку (`context.Canceled` или `context.DeadlineExceeded`):

```go
package main

import "context"

func WithContexCheck(ctx context.Context, action func()) {
	if action == nil || ctx.Err() != nil {
		return
	}
	action()
}

func main() {
	ctx := context.Background()
	WithContexCheck(ctx, func() {
		// do something
	})
}
```

> Не нужно пихать `ctx.Err()`/`ctx.Done()` проверки вообще везде подряд. Проверяйте контекст только там, где реально есть смысл прерваться: перед/во время похода в базу, HTTP-запроса, любой потенциально долгой блокирующей операции. Если у вас просто пара быстрых арифметических действий — контекст там ничего не решает, только шум добавляет.

## `WithTimeout` — пример

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func makeRequest(ctx context.Context) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		fmt.Println("finished")
	case <-ctx.Done():
		fmt.Println("canceled")
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	makeRequest(ctx) // напечатает "canceled" — таймаут в 1 сек наступит раньше, чем таймер в 5 сек
}
```

### ⚠️ Не вызвал `cancel()` — получил goroutine leak

Внутри `WithTimeout`/`WithDeadline` рантайм сам запускает служебную горутину-таймер, которая ждёт либо истечения времени, либо явного вызова `cancel()`. Если вы никогда не вызовете `cancel()` руками — эта горутина будет жить до самого дедлайна. Звучит не страшно, пока кто-нибудь не создаст контекст с `time.Hour * 24 * 365` (условным "годом") — тогда утечка станет вполне ощутимой.

**Правило простое: `defer cancel()` сразу же на следующей строке после создания контекста, без исключений.** Даже если кажется, что "он и так скоро сам отменится по таймауту" — `defer cancel()` ничего не стоит, а привычка спасает от реальных утечек.

## Своя реализация Context (для понимания, как это устроено внутри)

```go
package main

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Context struct {
	done   chan struct{}
	closed int32
}

func WithTimeout(parent Context, duration time.Duration) (*Context, func()) {
	if atomic.LoadInt32(&parent.closed) == 1 {
		return nil, nil // ⚠️ так делать не стоит, см. ниже
	}

	ctx := &Context{
		done: make(chan struct{}),
	}

	cancel := func() {
		// CompareAndSwap гарантирует, что close(ctx.done) выполнится РОВНО один раз,
		// даже если cancel() дёрнут одновременно из нескольких мест
		if atomic.CompareAndSwapInt32(&ctx.closed, 0, 1) {
			close(ctx.done)
		}
	}

	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-parent.Done(): // если родитель отменился раньше — отменяемся вместе с ним
		case <-timer.C: // а если нет — по своему таймауту
		}

		cancel()
	}()

	return ctx, cancel
}

func (c *Context) Done() <-chan struct{} {
	return c.done
}

func (c *Context) Err() error {
	select {
	case <-c.done:
		return errors.New("context deadline exceeded")
	default:
		return nil
	}
}

func (c *Context) Deadline() (time.Time, bool) {
	return time.Time{}, false // not implemented
}

func (c *Context) Value(any) any {
	return nil // not implemented
}

func main() {
	ctx, cancel := WithTimeout(Context{}, time.Second)
	defer cancel()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		fmt.Println("finished")
	case <-ctx.Done():
		fmt.Println("canceled")
	}
}
```

Зачем тут вообще `atomic.CompareAndSwapInt32` вместо простого `if !closed { close(done) }`: без атомарности это **гонка данных** — если `cancel()` дёрнут одновременно двумя горутинами (например, и явным вызовом, и по таймеру), обе могут одновременно увидеть `closed == 0` и обе попытаться сделать `close(ctx.done)` — а повторный `close` на уже закрытом канале паникует. CAS гарантирует, что "выиграет" только один вызывающий.

> ⚠️ Момент `return nil, nil` при уже отменённом родителе — плохая практика (это даже отмечено комментарием в исходном коде). Если кто-то дальше вызовет `ctx.Done()` на `nil`-указателе типа `*Context` — будет паника nil pointer dereference. Правильнее было бы вернуть уже отменённый контекст (например, с сразу закрытым `done`) или явно прокинуть ошибку родителя — но никак не голый `nil`.

## Дерево контекстов и каскадная отмена

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	makeRequest(ctx)
}

func makeRequest(ctx context.Context) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case <-newCtx.Done():
		fmt.Println("canceled")
	case <-timer.C:
		fmt.Println("timer")
	}
}
```

Логично было бы ожидать, что сработает либо `timer` (5 сек), либо собственный таймаут `newCtx` (10 сек) — то есть точно не раньше 5 секунд. Но вывод — `"canceled"` уже примерно через **2 секунды**.

Причина: `newCtx` создан **от** `ctx`, а родительский `ctx` отменяется через 2 секунды по своему таймауту. Отмена контекста **каскадно распространяется на все его дочерние контексты**, независимо от их собственных, более поздних дедлайнов. Дочерний контекст не может "пережить" родителя — максимум, на что он способен, это отмениться раньше родителя, но никак не позже.

```
                    ┌─────┐
                    │ ctx │  ── если отменить этот узел...
                    └──┬──┘
          ┌────────────┼──────────────┐
          │             │             │
      ┌─────┐        ┌─────┐        ┌─────┐
      │ ctx │        │ ctx │        │ ctx │
      └──┬──┘        └──┬──┘        └─────┘
         │              │
      ┌─────┐        ┌─────┐        ┌─────┐
      │ ctx │        │ ctx │        │ ctx │◄┐
      └──┬──┘        └──┬──┘        └──┬──┘ │
         │              │              │    │ ...отменятся ВСЕ узлы
       (конец)       ┌─────┐        ┌─────┐ │  во всём поддереве,
                     │ ctx │        │ ctx │◄┘  сколько бы у них ни было
                     └──┬──┘        └─────┘    своего таймаута
                        │
                     ┌─────┐
                     │ ctx │
                     └─────┘
```

Отсюда практический вывод: если где-то в глубине дерева вы даёте дочернему контексту больший таймаут, чем у родителя выше — это бессмысленно, родитель всё равно "срежет" его раньше. И это ещё один аргумент в пользу правила "отменяй там, где создал": отмена высоко в дереве бьёт по всем веткам ниже, даже тем, о существовании которых вы не думали в момент вызова `cancel()`.

## `WithValue` — как ищется значение

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	traceCtx := context.WithValue(context.Background(), "trace_id", "12-21-33")
	makeRequest(traceCtx)

	oldValue, ok := traceCtx.Value("trace_id").(string)
	if ok {
		fmt.Println("mainValue", oldValue)
	}
}

func makeRequest(ctx context.Context) {
	oldValue, ok := ctx.Value("trace_id").(string)
	if ok {
		fmt.Println("oldValue", oldValue)
	}

	newCtx := context.WithValue(ctx, "trace_id", "22-22-22")
	newValue, ok := newCtx.Value("trace_id").(string)
	if ok {
		fmt.Println("newValue", newValue)
	}
}
```

Тут многие путаются: кажется, что если переопределить значение по уже существующему ключу в дочернем контексте, оно "перезапишется" и в родителе тоже. **Это не так.** `WithValue` не мутирует родителя — она создаёт новый узел-обёртку, который просто **перекрывает** ключ для себя и всех _своих_ потомков. Родительский `traceCtx` как хранил `"12-21-33"`, так и продолжит хранить — `newCtx` живёт своей отдельной жизнью.

А вот обратная ситуация:

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	traceCtx := context.WithValue(context.Background(), "trace_id", "12-21-33")
	makeRequest(traceCtx)
}

func makeRequest(ctx context.Context) {
	oldValue, ok := ctx.Value("trace_id").(string)
	if ok {
		fmt.Println(oldValue)
	}

	newCtx, cancel := context.WithCancel(ctx) // оборачиваем в WithCancel, а не WithValue
	defer cancel()

	newValue, ok := newCtx.Value("trace_id").(string)
	if ok {
		fmt.Println(newValue) // найдётся ли значение родителя?
	}
}
```

Ответ — да, найдётся. `Value()` **рекурсивно поднимается вверх по дереву предков**, пока не найдёт совпадающий ключ или не дойдёт до корня. Не важно, через какую функцию (`WithCancel`, `WithTimeout`, `WithValue`...) создан промежуточный узел — если он сам не переопределяет этот ключ, поиск просто идёт выше, к родителю.

### Для чего вообще нужен `WithValue`

> Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions

На практике часто через мета-данные контекста пытаются протащить зависимости структуры, логгеры и подобное — просто чтобы не городить полноценный DI-контейнер. Формально не запрещено, но концептуально это не то, для чего задуман контекст, да ещё и добавляет накладные расходы (поиск по цепочке контекстов вплоть до корня, чтобы найти условный логгер родителя).

`WithValue` предназначен именно для **request-scoped** данных, которые естественным образом сопровождают запрос через границы функций/сервисов: `user_id`, `trace_id`, `request_id`, `txn_id` и подобное — то, что не является параметром бизнес-логики, а является метаданными про сам запрос.

### ⚠️ Коллизия ключей — баг, который стреляет редко, но метко

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	{
		ctx := context.WithValue(context.Background(), "key", "value1")
		ctx = context.WithValue(ctx, "key", "value2")

		fmt.Println("string =", ctx.Value("key").(string)) // "value2" — второй перекрыл первый
	}
	{
		type key1 string // отдельный именованный тип, не alias!
		type key2 string
		const k1 key1 = "key"
		const k2 key2 = "key"

		ctx := context.WithValue(context.Background(), k1, "value1")
		ctx = context.WithValue(ctx, k2, "value2")

		fmt.Println("key1 =", ctx.Value(k1).(string)) // "value1" — никто никого не перекрыл
		fmt.Println("key2 =", ctx.Value(k2).(string)) // "value2"
	}
}
```

В первом блоке мы **осознанно или случайно** перекрыли значение родителя тем же строковым ключом `"key"`. Проблема в том, что строковый литерал как ключ — это просто голая строка. Если в большом проекте два разных пакета (или два разных разработчика) независимо решат использовать `"trace_id"` как ключ — они молча начнут наступать друг другу на ноги, и это всплывёт не сразу.

Поэтому в Go community есть чёткая рекомендация: **используйте под ключ отдельный, желательно неэкспортируемый именованный тип.**

Почему это работает: ключ у `WithValue`/`Value` типа `any` (пустой интерфейс), а сравнение двух значений интерфейса в Go всегда сначала сравнивает **их динамический тип**, и только потом — значение. Поэтому `key1("key")` и `key2("key")` — это два совершенно разных ключа с точки зрения контекста, хотя строковое содержимое у них буквально одинаковое. Именно поэтому `type key1 string` — это обязательно **новый именованный тип** (`type key1 string`), а не alias (`type key1 = string`): alias был бы всё тем же `string` и коллизия вернулась бы.

> Две функции из разных, никак не связанных друг с другом пакетов вполне могут случайно использовать одинаковую строку в качестве ключа — собственный неэкспортируемый тип ключа полностью убирает этот риск.

## `WithCancelCause` — передаём причину отмены

```go
package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("error"))

	fmt.Println(ctx.Err())          // context.Canceled — общая, "родовая" ошибка
	fmt.Println(context.Cause(ctx)) // errors.New("error") — конкретная причина
}
```

Отдельно полезно, когда контекст могли отменить в нескольких местах дерева, и вам важно понять именно **кто и почему**. Если `cancel` с причиной вызвать несколько раз с разными ошибками — зафиксируется только **первый** вызов, все последующие ни на что не повлияют. Это ровно та же атомарность через CAS, что мы руками реализовывали в кастомном контексте выше — реальный `context` package устроен так же.

## `WithTimeoutCause`

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Second, errors.New("timeout"))
	defer cancel()

	<-ctx.Done()

	fmt.Println(ctx.Err())          // context.DeadlineExceeded
	fmt.Println(context.Cause(ctx)) // errors.New("timeout")
}
```

> ⚠️ Тонкость: `Cause` зависит от того, **что именно сработало первым**. Если контекст отменился по истечении таймаута — `Cause` будет той ошибкой, что вы передали в `WithTimeoutCause`. А если кто-то успел вызвать `cancel()` вручную раньше, чем истёк таймаут — `Cause` окажется стандартным `context.Canceled`, а не вашей кастомной причиной.

## `WithoutCancel`

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	innerCtx := context.WithoutCancel(ctx)
	cancel()

	if innerCtx.Err() != nil {
		fmt.Println("canceled")
	} else {
		fmt.Println("still alive") // напечатается это — innerCtx не почувствовал отмену родителя
	}
}
```

`WithoutCancel` возвращает копию родителя, которая **никогда не отменяется** сама (`Done()` возвращает `nil`-канал, `Err()` всегда `nil`), но при этом `Value()` по-прежнему прозрачно ходит вверх по дереву предков — то есть все значения (`trace_id`, `user_id` и т.д.) остаются доступны.

**Почему нельзя просто взять `context.Background()`**: тогда вы потеряете вообще всю связь с родителем — все его values исчезнут вместе с ним. `WithoutCancel` — это способ сказать: "мне не нужен сигнал отмены родителя, но мета-данные про исходный запрос мне всё ещё нужны". Типичный кейс — фоновая задача, которая должна пережить сам HTTP-запрос (например, дописать аудит-лог или отправить метрику уже после того, как ответ клиенту отдан и его контекст отменился), но при этом ей всё ещё хочется знать `trace_id` исходного запроса.

## Что делать, если библиотека не умеет в context

Бывает легаси-код или сторонняя либа, у функций которой просто нет параметра `context.Context` — и штатно отменить их вызов нельзя. Обходной путь — обернуть блокирующий вызов в горутину и гонять его наперегонки с `ctx.Done()`:

```go
package main

import (
	"context"
	"time"
)

func Query(s string) string { /* блокирующий вызов легаси-либы */ return s }

func DoQuery(queryStr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	resultCh := make(chan string, 1) // ⚠️ буфер обязателен, см. ниже
	go func() {
		result := Query(queryStr)
		resultCh <- result
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-resultCh:
		return result, nil
	}
}
```

**Почему канал обязательно буферизованный (хотя бы на 1):** если сработает `ctx.Done()` раньше, чем `Query` закончит работу, функция `DoQuery` вернёт управление и **никто больше никогда не прочитает** из `resultCh`. Если бы канал был unbuffered, горутина с `Query` дописала бы результат и намертво зависла бы на `resultCh <- result`, потому что читателя больше нет и никогда не будет — классический goroutine leak. С буфером на 1 отправка в канал успевает пройти без блокировки, горутина спокойно завершается и собирается GC, даже если результат никому уже не нужен.

> Важно понимать: сама легаси-функция `Query` физически продолжит выполняться до своего завершения — контекст не может "убить" уже запущенный код, который сам не проверяет отмену. Этот паттерн лишь освобождает **вызывающую сторону** раньше, а не прерывает саму операцию.

## Graceful Shutdown через `signal.NotifyContext`

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "hello world\n")
	})

	server := &http.Server{Addr: ":8888"}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Print(err.Error())
		}
	}()

	<-ctx.Done() // блокируемся тут, пока не придёт os.Interrupt (Ctrl+C)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() // ✅ даём серверу секунду на то, чтобы доработать активные запросы

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Print(err.Error())
	}

	fmt.Println("canceled")
}
```

> ⚠️ В исходном варианте конспекта сразу после создания контекста стоял вызов `cancel()` (а не `defer cancel()`), то есть контекст оказывался отменён **ещё до** вызова `server.Shutdown(ctx)`. Смысл в том, чтобы дать активным соединениям целую секунду на завершение — если контекст уже отменён в момент вызова `Shutdown`, он тут же форсированно оборвёт все соединения, не дожидаясь ничего. `defer cancel()` — единственный вариант, при котором таймаут реально работает как задумано.

`signal.NotifyContext` — удобная обёртка: превращает системный сигнал (`os.Interrupt`, то есть Ctrl+C) в обычный `context.Context`, который отменяется, как только сигнал придёт. Это избавляет от ручной возни с `signal.Notify` и каналом `os.Signal`.

## `context.AfterFunc`

Ещё одна функция пакета: регистрирует callback, который запустится **в отдельной горутине**, как только контекст завершится (по отмене или дедлайну):

```go
stop := context.AfterFunc(ctx, func() {
	fmt.Println("ctx завершён, чистим ресурсы")
})
defer stop() // если функция уже не нужна — можно отменить регистрацию через stop()
```

Удобно вместо того, чтобы каждый раз руками писать `go func() { <-ctx.Done(); ... }()`. Если `ctx` уже был завершён **до** вызова `AfterFunc` — callback всё равно запустится, просто немедленно (в новой горутине). `stop()` возвращает `bool`: `true`, если получилось отменить регистрацию до того, как callback стартовал, `false` — если callback уже запущен (или запускается).

---

## Рекомендации Go-community

- **Передавайте контекст всегда первым аргументом функции.** Это негласный, но абсолютно повсеместный конвенш в экосистеме Go — любой другой разработчик ожидает увидеть `ctx context.Context` первым параметром.
- **Передавайте только контекст, без функции отмены.** Если функция отмены "утекает" вниз по коду вместе с контекстом, контроль над жизненным циклом размывается — непонятно, кто в итоге отвечает за вызов `cancel()`.
- **Не храните контекст в структуре**, только передавайте его в функции/методы аргументом. Контекст спроектирован как одноразовый, неизменяемый объект, привязанный к конкретному вызову, а не как долгоживущее состояние.
- **`context.WithValue` — на крайний случай.** В подавляющем большинстве ситуаций нужные данные можно и нужно передать обычным аргументом функции — это и быстрее, и явнее, и не ломает поиск использований в IDE.
- **`context.Background()` — только на самом верху дерева**, как корень. Это просто заглушка без единого механизма контроля (`Done()` у неё никогда не закроется).
- **`context.TODO()` — если пока не уверены, какой контекст нужен.** Семантически идентичен `Background()`, но явно сигнализирует "тут ещё предстоит прокинуть нормальный контекст" — грепается отдельно и не путается с осознанным использованием `Background()`.
- **Никогда не передавайте `nil` вместо контекста.** Формально многие функции это даже не запретят на уровне компиляции, но по конвенции — если контекста реально нет, используйте `context.TODO()` или `context.Background()`, а не `nil`.