## What is a proof ?

A method of ascertaining truth. Meanwhile there are a lots of ways determining the truth, such as: 

- experiments
- sampling
- legal
- authority ? 
- religion
- inner conviction

But, as mathematicians we have own method of determining the truth, that is a **mathematical proof**.

A **mathematical proof** is a verification of a *proposition* by a chain of *logical deductions* from a base set of *axioms*.

## What is a proposition ?

A proposition is a statement that is true or false. Such as:

Proposition 1: Today is a Monday ? **True**

Proposition 2: 1 + 2 = 4 **False**

What about this, is this a proposition ? 

> For All *n* which belongs to natural number. n^2+n+41 is prime.

Even easier example, is this a proposition ?

>p is prime  

Not, cuz this is incomplete proposition, 12 is prime is composition, 5 is prime is proposition. But p is a variable that we should fill in to its become a proposition. But 'p is prime' is called **predicate**. 

## What is a predicate ?

A predicate is a proposition whose truth depends on variables. Kind of parameterized proposition. 

Is here ' For All *n* which belongs to natural number. n^2+n+41 is prime.' the 'n^2+n+41 is prime' is a predicate, and  ' For All *n* which belongs to natural number' fills or variables making the statement proposition. 

## Euler prime

All right then how we can identify that it is true ? The simplest way by manual iteration: 

| n   | n² + n + 41 | Prime? |
| --- | ----------- | ------ |
| 0   | 41          | ✓      |
| 1   | 43          | ✓      |
| 2   | 47          | ✓      |
| ⋮   | ⋮           | ⋮      |
| 39  | 1601        | ✓      |

For some field of study doing a sufficient number of examples would be enough  to say it is true. Unfortunately for us mathematicians for statement for ALL n just this forty examples is not enough. 

So when we try example with 40, 41, 42 and etc we get not prime numbers. So this proposition is **False**.

## Goldbach's Conjecture

> Everyu even number > 2 is the sum of 2 primes. 

Lets start by exampling, we take 12 ? 

> 12 = 7 + 5
> 22 = 17 + 5 

And seems this is true, but we could not proof it, cuz it is the numbers theory. This is a part of nature. And that way it called **conjecture**. 

Even though we can express 20 as 11  + 9 (9 is not prime), at the same time we can express it as 17 + 3 :)

## Truth Table

A, B are proposition. 

Not A - It means A is **Flase**, we can build truth table for this logical statement:

| A   | not A |
| --- | ----- |
| T   | F     |
| F   | T     |

NOT (notations: ) - 

And we can take two proposition and combine them in one truth table, which gives as already 4 states:


| A   | B   | A and B | A or B |
| --- | --- | ------- | ------ |
| T   | T   | T       | T      |
| T   | F   | F       | T      |
| F   | T   | F       | T      |
| F   | F   | F       | F      |

AND (notations: ) -

(inclusive) OR (notations: )  -

In real life our OR statements works a bit differently some times, imagine situation, I have been making a soup, and your mom comes to you and ask did you put pasta or chicken into soup. The possible answers could be "chicked", "pasta", but not really both (cuz we doing eather pasta soup or chicken soup), and nor neither (cuz one should added its to be soup)

and that's is operand XOR.

then mom ask do you want coffee or tea. You can ask for coffee, tea, ~~both~~, neither. And this or in mathematician called NAND (not of AND).

then mom ask do you want cream or sugar ? cream, sugar, both, neither. Means all of the options are true. but this or in math is just true ahaha :) 

## Implication

A implies B (notation:  A -> B)

A implies B means, "if A then B".

Now it more like what computer language does, lets see the truth table of implication


| A   | B   | A -> B |
| --- | --- | ------ |
| T   | T   | T      |
| T   | F   | F      |
| F   | T   | ?      |
| F   | F   | T      |
If A is true then B is true. Thats is true statements in terms of implication 

If A is true then B is false. Thats is false statements in terms of implication, agree it does not means if A then B. 

IF A is false then B is false, Thats true as well. 

But the trickiest part if A is false then B is true. What do you think it is seems false or true intuitively ? To understand all of this we need concrete example: 

> On Wednesdays we wear pink

If Wednesday then pink (Wed -> pink)

| Wed | Pink | Obeying the rule ? |
| --- | ---- | ------------------ |
| T   | T    | Yes (T)            |
| T   | F    | No (F)             |
| F   | T    | Yes (T)            |
| F   | F    | Yes (T)            |
Thats makes more sense!

Even though implication is complicated and confusing. Lets explore more. In English when we say A implies B, we mean causality = A causes B, and also means time = A happens and after B happens. But we in math implication we dont care about it, as mathematicians we only care Truth Tabler we had and only it. 

A -> B not equal notA -> notB

Cuz If not Wednesday, then no pink. Obeys the pink on others days. It is not saying U can wear pink in other days, which did A -> B. 

But this is called inverse.

As well as here  