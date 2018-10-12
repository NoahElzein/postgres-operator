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

## Experiment Resources
* AWS EC2 Instance - We will run the modified Kubernetes cluster here.
* [Kubernetes Setup Guide](https://dzone.com/articles/easy-step-by-step-local-kubernetes-source-code-cha)
* [PostgresController](https://github.com/NoahElzein/postgrescontroller)

## Experiment General Steps
* Setup AWS EC2 Instance.
* Setup Kubernetes on EC2 instance and make sure it works.
* Modify Kubernetes to allow for API call detection.
* Clone Postgres Operator repo.
* Make sure all dependencies are satisfied.
* Begin experiment.

## Setup AWS EC2 Instance
Go to AWS and create an EC2 instance with these specifications:
- Ubuntu 18.04 image
- m5d.xlarge instance

## Setup Kubernetes on EC2 Instance and make sure it works
Follow the article referred to above for detailed steps in initializing Kubernetes on your AWS instance.
[Link Here](https://dzone.com/articles/easy-step-by-step-local-kubernetes-source-code-cha)

## Modify Kubernetes to allow for API call detection
Turn off Kubernetes instance if currently running.
Navigate to kubernetes/staging/src/k8s.io/apiserver/pkg/server/filters/
Open the wrap.go file.
Find this line:
```
logger := httplog.NewLogged(req, &w)
```
Right before the above line, insert this code: 
```
urlPath := req.URL.String()

if strings.Contains(urlPath, "crontab") {
      glog.Infof("==========================================================")
      glog.Infof("Hello hello crontabs")                
      glog.Infof("==========================================================")
}
```

Note: Do not forget to import the strings library into wrap.go.

## Clone Postgres Operator repo
After setting up Kubernetes, we need to setup our repo. The guide in the article also deals with setting up the $GOPATH
so a lot of the work is done thankfully. Clone the repo using: 

```
git clone https://github.com/NoahElzein/postgrescontroller
```

We need to install two dependencies for this repo. 

[lib/pq](https://github.com/lib/pq)

Clone the above repo under the guthub.com folder.

```
git clone https://github.com/lib/pq.git
```

We also need to install the Postgres command line tool:

```
sudo apt-get install postgresql postgresql-contrib
```

## Begin Experiment

Run the Kubernetes cluster.

Run postgrescontroller in a separate terminal window.

Access the tmp folder and tail the apiserver logs:

```
tail -f kube-apiserver.log
```

Open another terminal window

Go into the postgrescontroller project

Register a Postgres object using the kubectl command:

```
kubectl create -f deploy/crd.yaml
```

You should see the message we inserted in wrap.go pop up in the log file. 

Now it is time to begin the experiment:

The way we will measure performance of the library is based on the number of API calls that appear over a period of time during experimentation. The reason for this is because at the moment there is no way to differentiate between Resync API calls and API calls resulting from Library calls in Postgrescontroller.

The format I followed was a 5 minute interval involving creation of DB as well as Users in the Postgrescontroller.

Minute 0: Turn on Kubernetes cluster and record calls.

Minute 2: Create this resource:

```
kubectl create -f deploy/initializeclient.yaml
```

Minute 3:
```
kubectl create -f deploy/add-db.yaml
```

Minute 4: 

```
kubectl create -f deploy/add-user.yaml
```

Minute 5: Conclude experiment.

We can run the same experiment on the custom controller.

Clone this repository [Custom Controller](https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd):

```
git clone https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd.git
```

Follow the steps outlined in the Github repo. No major changes are required. 

Perform the same experiment on the Custom Controller.

## Results

The custom controller resulted in 37-39 API calls when performed on average.

The Postgres Operator resulted in 54-56 API calls when perfomed on average.


