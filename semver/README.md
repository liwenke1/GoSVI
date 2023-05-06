# Introduction

This repository is the implementation and data storage of paper [A Large-Scale Empirical Study on Semantic Version in Go Open Source Ecosystem]()

## Requirements

* Go 1.19
* Python3
* arrow
* tqdm
* wordcloud
* crawl

## Installation

* clone this repository

* download [complete go repository](https://drive.google.com/file/d/1Y3x01pw9vaspQv37EXCJD_Rxj9-xF-kN/view?usp=sharing) to replace [mini file](results/dataset/GoRepositoryMini.txt)

* download [complete analysis result](https://drive.google.com/file/d/1oRmvOrtkEAak962ERnm7F7ZOB2QplQK8/view?usp=sharing) about stable and unstable version information to replace [mini file](results/data/LibraryVersionStableAndUnstableNameMini.json)


## Usage

* use [crawl](crawl/) module to collect go repository in github and their dependency relationship

* use [api diff](semver/) module to detect semver breaking

* use [scripts](scripts/) module to extract all kinds of data from detection results, we give some results [here](results/)
