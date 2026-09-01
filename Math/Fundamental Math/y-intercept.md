Now you know that [[slope]] identifies the steepness of the line graph. But we considered [[equations]], only going through [[Two Variable Equations#^theorigin|the origin]], let's explore what happens when it does not:
```functionplot
---
title: 
xLabel: x
yLabel: y
bounds: [-5,5,-5,5]
disableZoom: false
grid: true
---
y = 2x  + 4
```

Let's calculate the slope for this graph, considering these $P_{1}=(-1;2)$ and $P_{2}=(0;4)$:
$$\displaylines{
m = \frac{4-2}{0-(-1)} \\
m = \frac{2}{1} \\
m=2
}$$
So, according to previous [[Slope|slope rule]] the equation of the graph should look like this: 
$$y = 2x$$
Let's test our first point x=-1 and y=2:
$$\displaylines{
y = 2 \times -1 \\
y = -2 \\
2 \neq -2
}$$
 As you can see, something is missing here, and what we're missing is the consideration of the **y-intercept** change. Previously, when we explored the concept of slope, we considered graphs that pass through the origin, represented by the coordinates (0, 0), which implies a y-intercept of 0. Therefore, our initial equation was presented as y = 2x, which indeed is equivalent to y = 2x + 0, highlighting the y-intercept's role. So, the right equation will be, first identifying the y-intercept which is equal to 4 and adding it to the equation:
$$y = 2x + 4$$

Let's check substituting first point x=-1 and y = 2:
$$\displaylines{
y = 2 \times (-1) + 4\\
y = (-2) + 4 \\
y = 2 \\
2 = 2
}$$

In conclusion, We can write the equation of any line given its [[slope]] and **y -intercept**. 
>[! Line Graph]
>A line with slope $m$ and a *y*-intercept of $(0,b)$ can be written as: $$y=mx+b$$
