So far we have explored so many [[Equations|equation]] graphs go on forever at least in one direction. Now, let's delve into how mathematical language can depict various other graph forms.

## A Bounded Graphs 
___
$$x^{2}+y^{2}=25$$

Let's see how variables are related in table form:

| x   | y     |
| --- | ----- |
| 0   | 5, -5 |
| 5   | 0     |
| -5  | 0     |
| 3   | 4, -4 |
| 4   | 3, -3 |
| -3  | 4, -4 |
| -4  | 3. -3 |

Let's see how is the graph look like:

```functionplot
---
title: 
xLabel: 
yLabel: 
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y = -sqrt(25 - x^2)
y = +sqrt(25 - x^2)
```

As you can see, we representing a Circle with equations. Both axis has [[Minimums and Maximums#Absolute maximum|absolute maximum]] and [[Minimums and Maximums#Absolute minimum|absolute minimum]]. It's a bounded graph.  

## Impossible Points
$$x^{2}y=8$$
We've seen lots of completely continuous graphs — let's build one that has a break in it.

Let's see how variables are related in table form:

| x   | y   |
| --- | --- |
| 1   | 8   |
| 2   | 2   |
| 4   | 0,5 |
| -1  | 8   |
| -2  | 2   |
| -4  | 0.5 |
Let's see explore it's graph:
```functionplot
---
title: 
xLabel: 
yLabel: 
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=8/x^2
```
Is there any absolute minimum or maximum, let's solve equation substituting zero:

When x = 0:
$$\displaylines{
0^{2}y=8 \\
0y=8 \\
0=8
}$$
When y = 0:
$$\displaylines{
x^{2}0=8 \\
0=8 \\
}$$

Zero times **any** number is zero, so we get an equation that can't be true no matter the value of y or x.

So, for $x^{2}y=8$, the graph never reaches an x-coordinate **or** y-coordinate of zero. Subsequently, the equation does not have any absolute maximum or minimum. 

