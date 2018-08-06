package stub

import (
	"os"
	"os/exec"
	"io/ioutil"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/demo/postgrescontroller/pkg/apis/postgrescontroller/v1"

		"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"fmt"
	"time"
)

const controllerAgentName = "postgres-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

const (
      PGPASSWORD = "mysecretpassword"
      MINIKUBE_IP = "192.168.99.100"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.Postgres:
		foo := o
		deploymentName := foo.Spec.DeploymentName

		if deploymentName == "" {
			logrus.Errorf("Deployment name must be specified")
			return nil
		}

		var verifyCmd string
		var actionHistory []string
		var serviceIP string
		var servicePort string
		var setupCommands []string
		var err error

		serviceIP, servicePort, setupCommands, verifyCmd, err = createDeployment(foo)

		if err != nil {
			/* TO DO: If it already exists, we are currently returning null values.
			   That is bad because we cannot update with null values. We need the actual values.
			   I think the way to do this would be to retrieve the fields in the deployment.
			   For some reason we are retrieving the fields in the Postgres object as shown below
			   That information won't be stored there as far as I know */
			fmt.Println("THE ERROR OF CREATE DEPLOYMENT IS NOT NULL")
			if apierrors.IsAlreadyExists(err) {

				fmt.Println("THE RESOURCE ALREADY EXISTS BASED ON WHAT IS BEING SAID HERE.")
				actionHistory := foo.Status.ActionHistory
				serviceIP := foo.Status.ServiceIP
				servicePort := foo.Status.ServicePort
				verifyCmd := foo.Status.VerifyCmd
				fmt.Printf("Action History:[%s]\n", actionHistory)
				fmt.Printf("Service IP:[%s]\n", serviceIP)
				fmt.Printf("Service Port:[%s]\n", servicePort)
				fmt.Printf("Verify cmd: %v\n", verifyCmd)

				setupCommands = canonicalize(foo.Spec.Commands)

				var commandsToRun []string
				commandsToRun = getCommandsToRun(actionHistory, setupCommands)
				fmt.Printf("commandsToRun: %v\n", commandsToRun)

				if len(commandsToRun) > 0 {
					err2 := updateFooStatus(foo, &actionHistory, verifyCmd, serviceIP, servicePort, "UPDATING")
					if err2 != nil {
						return err
					}
					updateCRD(foo, commandsToRun)
					for _, cmds := range commandsToRun {
						actionHistory = append(actionHistory, cmds)
					}
					err2 = updateFooStatus(foo, &actionHistory, verifyCmd,
						serviceIP, servicePort, "READY")
					if err2 != nil {
						return err
					}
				}

			} else {
				panic(err)
			}
		} else {
			fmt.Println("THE ERROR IS NULL SO GOOD NEWS.")
			for _, cmds := range setupCommands {
				actionHistory = append(actionHistory, cmds)
			}
			fmt.Printf("Setup Commands: %v\n", setupCommands)
			fmt.Printf("Verify using: %v\n", verifyCmd)

			fmt.Println("BEFORE THE FOO STATUS UPDATE")
			err1 := updateFooStatus(foo, &actionHistory, verifyCmd, serviceIP, servicePort, "READY")
			fmt.Println("I CAME BACK FROM THE UPDATEFOOSTATUS")

			if err1 != nil {
				return err1
			}
		}

		/*err := sdk.Create(newbusyBoxPod(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create busybox pod : %v", err)
			return err
		} */

	}

	return nil
}

func updateCRD(foo *v1.Postgres, setupCommands []string) {
	serviceIP := foo.Status.ServiceIP
	servicePort := foo.Status.ServicePort

	//setupCommands1 := foo.Spec.Commands
	//var setupCommands []string

	//Convert setupCommands to Lower case
	//for _, cmd := range setupCommands1 {
	//	 setupCommands = append(setupCommands, strings.ToLower(cmd))
	//}

	fmt.Printf("Service IP:[%s]\n", serviceIP)
	fmt.Printf("Service Port:[%s]\n", servicePort)
	fmt.Printf("Command:[%s]\n", setupCommands)

	if len(setupCommands) > 0 {
		file := createTempDBFile(setupCommands)
		fmt.Println("Now setting up the database")
		setupDatabase(serviceIP, servicePort, file)
	}
}

func getCommandsToRun(actionHistory []string, setupCommands []string) []string {
	var commandsToRun []string
	for _, v := range setupCommands {
		var found bool = false
		for _, v1 := range actionHistory {
			if v == v1 {
				found = true
			}
		}
		if !found {
			commandsToRun = append(commandsToRun, v)
		}
	}
	fmt.Printf("-- commandsToRun: %v--\n", commandsToRun)
	return commandsToRun
}

func updateFooStatus(foo *v1.Postgres,
	actionHistory *[]string, verifyCmd string, serviceIP string, servicePort string,
	status string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	fooCopy := foo.DeepCopy()
	//fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	fooCopy.Status.AvailableReplicas = 1

	//fooCopy.Status.ActionHistory = strings.Join(*actionHistory, " ")
	fooCopy.Status.VerifyCmd = verifyCmd
	fooCopy.Status.ActionHistory = *actionHistory
	fooCopy.Status.ServiceIP = serviceIP
	fooCopy.Status.ServicePort = servicePort
	fooCopy.Status.Status = status
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	err := sdk.Update(fooCopy)
	return err
}

func createDeployment (foo *v1.Postgres) (string, string, []string, string, error){

	deploymentName := foo.Spec.DeploymentName
	image := foo.Spec.Image
	username := foo.Spec.Username
	password := foo.Spec.Password
	database := foo.Spec.Database
	setupCommands := canonicalize(foo.Spec.Commands)

	fmt.Printf("   Deployment:%v, Image:%v, User:%v\n", deploymentName, image, username)
	fmt.Printf("   Password:%v, Database:%v\n", password, database)
	fmt.Printf("   SetupCmds:%v\n", setupCommands)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},

				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  deploymentName,
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 5432,
								},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									TCPSocket: &apiv1.TCPSocketAction{
										Port: apiutil.FromInt(5432),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds: 60,
								PeriodSeconds: 2,
							},
							Env: []apiv1.EnvVar{
								{
									Name: "POSTGRES_PASSWORD",
									Value: PGPASSWORD,
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Creating deployment...")
	err := sdk.Create(deployment)
	fmt.Println("Before the error is supposed to happen")
	fmt.Println(err)
	if err != nil {
		fmt.Printf("I am about to return! This does not make any sense...")
		return "", "", nil, "", err
	}
	fmt.Println("After the error is supposed to happen")

	fmt.Printf("Created deployment %q.\n", deployment.GetObjectMeta().GetName())
	fmt.Printf("------------------------------\n")

	fmt.Printf("Creating service.....\n")

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Namespace: "default",
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort {
				{
					Name: "my-port",
					Port: 5432,
					TargetPort: apiutil.FromInt(5432),
					Protocol: apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string {
				"app": deploymentName,
			},
			Type: apiv1.ServiceTypeNodePort,
		},
	}

	fmt.Println("I am next to the service creation...")
	err2 := sdk.Get(service)
	if err2 != nil {
		fmt.Println(err2)
		fmt.Println("I could not find the service, this is a good thing, I guess...")
	}
	err1 := sdk.Create(service)

	if err1 != nil {
		fmt.Println("THE SERVICE IS ALSO FAILING...")
		return "", "", nil, "", err
	}

	fmt.Printf("Created service %q.\n", service.GetObjectMeta().GetName())
	fmt.Printf("------------------------------\n")

	serviceIP := MINIKUBE_IP

	nodePort1 := service.Spec.Ports[0].NodePort
	nodePort := fmt.Sprint(nodePort1)
	servicePort := nodePort

	time.Sleep(time.Second * 5)

	for {
		readyPods := 0
		pods := getPods()

		for _, d := range pods.Items {
			podConditions := d.Status.Conditions
			for _, podCond := range podConditions {
				if podCond.Type == corev1.PodReady {
					if podCond.Status == corev1.ConditionTrue {
						readyPods += 1
					}
				}
			}
		}

		if readyPods >= len(pods.Items) {
			break
		} else {
			fmt.Println("Waiting for Pod to get ready.")
			time.Sleep(time.Second * 4)
		}
	}

	time.Sleep(time.Second * 2)

	if len(setupCommands) > 0 {
		file := createTempDBFile(setupCommands)
		fmt.Println("Now setting up the database")
		setupDatabase(serviceIP, servicePort, file)
	}

	verifyCmd := strings.Fields("psql -h " + serviceIP + " -p " + nodePort + " -U <user> " + " -d <db-name>")
	var verifyCmdString = strings.Join(verifyCmd, " ")
	fmt.Printf("VerifyCmd: %v\n", verifyCmd)
	return serviceIP, servicePort, setupCommands, verifyCmdString, nil

}

func setupDatabase(serviceIP string, servicePort string, file *os.File) {
	defer os.Remove(file.Name())
	args := strings.Fields("psql -h " + serviceIP + " -p " + servicePort + " -U postgres " + " -f " + file.Name())
	fmt.Printf("Database setup command: %v\n", args)

	envName := "PGPASSWORD"
	envValue := PGPASSWORD
	newEnv := append(os.Environ(), fmt.Sprintf("%s=%s", envName, envValue))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = newEnv

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Err:%v\n", err)
		fmt.Printf("Out:%v\n", out)
		fmt.Printf("Out:%s\n", out)
		panic(err)
	}
}

func createTempDBFile(setupCommands []string) (*os.File){
	file, err := ioutil.TempFile("/tmp", "create-db1")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Database setup file:%s\n", file.Name())

	for _, command := range setupCommands {
		//fmt.Printf("Command: %v\n", command)
		// TODO: Interpolation of variables
		file.WriteString(command)
		file.WriteString("\n")
	}
	file.Sync()
	file.Close()
	return file
}

// podList returns a v1.PodList object
func getPods() *apiv1.PodList {
	return &apiv1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

func canonicalize(setupCommands1 []string) []string {
	var setupCommands []string

	for _, cmd := range setupCommands1 {
		setupCommands = append(setupCommands, strings.ToLower(cmd))
	}

	return setupCommands
}

func int32Ptr(i int32) *int32 { return &i }

// newbusyBoxPod demonstrates how to create a busybox pod
func newbusyBoxPod(cr *v1.Postgres) *corev1.Pod {
	labels := map[string]string{
		"app": "busy-box",
	}
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "busy-box",
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1.SchemeGroupVersion.Group,
					Version: v1.SchemeGroupVersion.Version,
					Kind:    "Postgres",
				}),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
