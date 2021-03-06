
vertice
=======

Vertice is the core engine for Megam Vertice 1.5.x and is open source. 


# Roadmap for 2.0

Read the [Deployment design](https://github.com/megamsys/verticedev/blob/master/proposals/01.deployments.md).

## Where is the code for 2.0

We have moved the development to private gitlab as it will have enterprise features.

## When can i get it in my anxious hands

`2.0` will be released on `Sep 30 2017` or less.

`2.0` developed with private enterprise features and is moved to gitlab.


### Requirements

>
[Golang 1.8 > +](http://www.golang.org/dl)

>
[NSQ 0.3.x](http://nsq.io/deployment/installing.html)

>
[Cassandra 3 +](https://wiki.apache.org/cassandra/GettingStarted)

## Usage

``vertice -v start``


### Compile from source


```
mkdir -p code/megam/go/src/github.com/megamsys

cd code/megam/go/src/github.com/megamsys

git clone https://github.com/megamsys/vertice.git

cd vertice

make

```


### Documentation

[development documentation] (https://github.com/megamsys/verticedev/tree/master/development)

[documentation] (http://docs.megam.io) for usage.




We are glad to help if you have questions, or request for new features..

[twitter @megamsys](http://twitter.com/megamsys) [email support@megam.io](<support@megam.io>)

[devkit] (https://github.com/megamsys/verticedev)

# License


|                      |                                          |
|:---------------------|:-----------------------------------------|
| **Author:**          | Rajthilak (<rajthilak@megam.io>)
| 	                   | KishorekumarNeelamegam (<nkishore@megam.io>)
|                      | Ranjitha  (<ranjithar@megam.io>)
|                      | MVijaykanth  (<mvijaykanth@megam.io>)
| **Copyright:**       | Copyright (c) 2013-2017 Megam Systems.
| **License:**         | Apache License, Version 2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
