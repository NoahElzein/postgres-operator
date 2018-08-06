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

After running the Minikube cluster, go into pkg/stub/handler.go and modify the variable MINIKUBE_IP to match the IP address
of your Minikube cluster.

To find the IP address of your Minikube cluster run:

```
$ minikube ip
```

Before running the Operator make sure the CRD is applied to the cluster as a supported API resource:

```
$ kubectl create -f deploy/crd.yaml
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
$ kubectl create -f deploy/client1.yaml
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

## Acknowledgements
This Postgres operator is completely based off of Cloud-Ark's [Postgres Custom Controller](https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd) as is most of the code in the Handler.go file. 
