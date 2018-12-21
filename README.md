# postgres-operator
PostgreSQL Operator generated using Operator SDK library.

## Operator SDK
This project makes use of the Operator SDK library to create a Kubernetes operator that
will handle and watch the creation of Postgres resources.

## Requirements

* [Minikube](https://kubernetes.io/docs/setup/minikube/) - For testing on a local cluster.
* kubectl - To interact with Kubernetes cluster.
* [golang](https://golang.org/dl/) - Language used for Operator development and deployment.
* PostgreSQL - Execute SQL queries locally that interact with the Postgres deployment.

## How to test:

After running the Minikube cluster, go pkg/controller/postgres/postgres_controller.go and modify the variable MINIKUBE_IP to match the IP address
of your Minikube cluster.

To find the IP address of your Minikube cluster run:

```
$ minikube ip
```

Before running the Operator make sure the CRD is applied to the cluster as a supported API resource:

```
$ kubectl create -f deploy/crds/postgresoperatorsdknew_v1_postgres_crd.yaml
```

Run the Operator locally:

```
$ operator-sdk up local --kubeconfig=$HOME/.kube/config
```

NOTE: The flag --kubeconfig redirects the operator to be configured to interact with minikube. 
      The handler needs to know where minikube is and how to execute functions to interact with 
      minikube. The above path is the default path to the .kube/config. If for whatever reason
      your path is different, make the necessary changes to match your path.

Open a new terminal while the Operator is running in the other one.

Run the following command to begin testing:

```
$ kubectl create -f client1.yaml
```

The above command creates a Postgres instance named client1 that contains a Postgres image 
and several Postgres queries to be ran upon launch of the instance.
The Postgres Operator should catch this deployment and begin creating the corresponding Deployment and Service objects.

Run the following commands to verify that the operator is doing what it is supposed to do:

```
$ kubectl get crd
$ kubectl get postgres client1
$ kubectl describe postgres client1
$ psql -h <IP> -p <port> -U <username> -d <db-name>
```
For the last command above, plug in the values of the IP and port numbers generated in the Operator output
as per service object availability.
For the username and db-name, go to deploy/client1.yaml and use the values for the username and database specified.
When prompted for password, use the password in the client1.yaml as well.

# Postgres Kubebuilder

## Kubebuilder
The Kubebuilder project involves the same Postgres program, only it is implemented using kuebuilder.

The requirements are the exact same as Operator SDK so we will skip to how to test.

## How to test:

While running minikube run:

```
$ make install
$ make run
```

This will run the Kubebuilder Postgres, after that follow the same steps from Operator SDK to test that the Postgres server is up and running.

# Building an Operator SDK program:

These are general steps of creating an Operator that will then be compared to kubebuilder steps for creating a skeleton controller.

Start with creating a skeleton project:

```
$ operator-sdk new postgres-operator-sdk-new
```

Next step is to create an API (Your Custom Resource):

```
$ operator-sdk add api --api-version=postgresoperatorsdknew.kubeplus/v1 --kind=Postgres
```

This will generate the custom resource under the pkg/apis directory. 

We also need to create a controller that will watch this API:

```
$ operator-sdk add controller --api-version=postgresoperatorsdknew.kubeplus/v1 --kind=Postgres
```

This will generate a controller skeleton under pkg/controller.

NOTE: Whenever making changes to _types.go, you must regenerate code for the resource:

```
$ operator-sdk generate k8s
```

From there on. We have a directory skeleton. Our main modifications will occur within pkg/apis (Making modifications to <RESOURCE_NAME_type.go>) as well as under pkg/controller/ directory where we modify the controller.

# Building a Kubebuilder program: 

Create a project skeleton using:

```
$ kubebuilder init
```

This will generate a very similar directory structure as Operator SDK

Next generate the API's:

```
$ kubebuilder create api --group kubeplus --version v1 --kind Postgres
```

This will generate the API under pkg/api

It will also automatically generator the controller for that API under pkg/controller.

Make modifications to pkg/apis and pkg/controller then run "make" to regenerate the code.

# Comparison of writing a Controller with kubebuilder vs with Operator SDK:

Both frameworks rely are built on top of controller-runtime library, this means that not only are the skeleton directories very similar, the process of writing the controller is almost identical as well.

Both frameworks rely on a manager that initializes the controller before it runs. Under the pkg/controller/ directory you can find your generated controller with a Reconcile function. This is the function that dictates the action of your controller when changes are made to your CRD.

Both frameworks go through the steps of generating a skeleton, then creating the appropriate API and Controller. 

.............
.............


## Acknowledgements
This Postgres operator is completely based off of Cloud-Ark's [Postgres Custom Controller](https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd) as is most of the code in the Handler.go file.
