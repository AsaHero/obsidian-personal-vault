So far we explore problems having one unknown factor. In real life we encounter problems with multiple unknown values. Math allows represent them too. 

Let's start exploring [[Equations|equations]] with two variables. For example: 
$$2y+5=x+5$$

In this kind of equation's [[Expressions#Variables|variables]] are related to each other, so when one variable changes, it will effect to other as well. Subsequently, these equations will have set of solutions rather than one solution.

>[!Note!]
>Two Variable Equations have unlimited number of solutions.


First method of solving two variable equations, testing values. Test substituting a random value to one of the variable and analyze how other relates. Let's see table representation of the set of solutions:

$2y+5=x+5$

| x   | y   |
| --- | --- |
| 2   | 1   |
| 4   | 2   |
| 6   | 3   |

#### Cartesian Plane
___
The set of solution can be represented as coordinates in the **Cartesian Plane**, having x and y coordinates.

This is a **Cartesian Plane**:
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: false
grid: true
---

```

The vertical and horizontal lines in the center of this Cartesian Plane called axis.
- Vertical line called: **y-axis**.
- Horizontal line called **x-axis**.

> [! The Origin point]
>   The point (0,0) is called **the origin**. It is the point where the the point where the x -axis and y -axis intersect. ^theorigin

As we known there are infinity many solutions in two variable equations, so to show that there are infinite solutions between points, we plot a **continuous** graph. Let's plot more two variable equations using testing values, to explore graph patterns on Cartesian plane. 

1. $y = 2x + 5$

```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y = 2x+5
```


2.  $y = \frac{8}x -4$ 
```functionplot
---
title:
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=8/x-5
```

3. $y=\sqrt x$

```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: false
grid: true
---
y=sqrt(x)
```




4. $3y = 9 + x^2$
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-20,20,-20,20]
disableZoom: true
grid: true
---
f(x) = (x^2  + 9)/3
```

5. $y = x^3$

```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y = x^3
```

6. $y = |x|$ 

> [! Absolute Value]
> The **absolute value** of a number is the distance of the number from zero. For example, ∣5∣=5 and ∣−10∣=10. ^absolute-value

```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-10,10,-10,10]
disableZoom: true
grid: true
---
y=abs(x)
```
