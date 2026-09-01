### Basic Schema

Uses labeled inputs (meaning the input has a corresponding output label) to train models and learn outputs.

![[Pasted image 20241123231304.png]]

The simplest schema of supervised learning involves:

- **Input Data**: A collection of features (**features** vector).
- **[[Model]]**: Learns the relationship between input features and output labels.
- **Output**: Predicted labels or values based on the trained model.

![[Pasted image 20241123232614.png]]


## Features (Input Data)
---

A **features vector** is a collection of data points or attributes used to train a supervised learning model. Features can be broadly categorized into **qualitative** and **quantitative** data types.
### Qualitative Features

Qualitative data is **categorical**, meaning it represents distinct categories or groups.

Examples:

- Gender (male, female)
- Continents
- Nationality
- Location
- Oceans

Qualitative data can be further classified as:

#### 1. **Nominal Data**

- Data without inherent order.
- Example: Gender (male, female).

#### 2. **Ordinal Data**

- Data with inherent order or ranking.
- Example: Age groups (baby, toddler, adolescent, adult).

> [!Nominal vs Ordinal Data]
> 
> - **Nominal**: No order or ranking (e.g., gender, continents).
> - **Ordinal**: Inherent order (e.g., age groups: baby → toddler → adolescent → adult or reactions: bad -> good -> great).
> 
> ![[Pasted image 20241123234459.png]]
### Quantitative Features

Quantitative data is **numerical**, representing measurable values.

Examples:

- Length (e.g., in meters)
- Temperature (e.g., in Celsius)
- Remaining eggs in a fridge

Quantitative data can be further divided into:

#### 1. **Discrete Data**

- Finite or countable values.
- Example: Number of eggs in a fridge (e.g., 12, 11, 10).

#### 2. **Continuous Data**

- Infinite possible values within a range.
- Example: Temperature (e.g., 22.5°C, 22.6°C).


## Tasks (Output Data)
---
Supervised learning models are trained to predict output data, which can be used to accomplish various tasks. One common task is **classification**, where the goal is to assign input features to specific categories.
### [[Classification|Classification Task]]

**Classification** involves predicting **discrete classes** based on input features. There are two main types of classification:

#### 1. **Multi-class Classification**

- The model predicts from more than two possible categories.
- Example:
    - Given features of food items, classify them as **hot-dog**, **pizza**, or **ice-cream**.
    - New features provided to the model will be labeled with one of the categories.

![[Pasted image 20241124011353.png]]
#### 2. **Binary Classification**

- The model predicts between only two possible categories.
- Example:
    - Determine if an item is **a hot-dog** or **not a hot-dog**.
    - Other common examples:
        - **Spam or not spam** emails.
        - **Pass or fail** outcomes.

![[Pasted image 20241124011823.png]]

### Regression Task

Predict continuous values. It will predict some continuous number, such as price of bitcoin, the value of temperature, the price of house and etc. 

![[Pasted image 20241124012545.png]]

## Diagram: Classification Example

Below is a conceptual representation of classification tasks:

|**Task Type**|**Input Features**|**Predicted Output**|
|---|---|---|
|**Multi-class**|Food features (shape, size)|Hot-dog, Pizza, Ice-cream|
|**Binary**|Email content (words, links)|Spam, Not Spam|
