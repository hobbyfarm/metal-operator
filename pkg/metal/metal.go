package metal

import (
	"context"
	"fmt"
	"strings"

	equinixv1alpha1 "github.com/hobbyfarm/metal-operator/pkg/api/v1alpha1"
	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MetalClient struct {
	*packngo.Client
	ProjectID string
}

const (
	ReservationAnnotation = "elasticReservationID"
	AddressAnnotation     = "elasticIP"
)

func NewClient(ctx context.Context, client client.Client, secret string, namespace string) (m *MetalClient, err error) {
	m = &MetalClient{}
	credSecret := &corev1.Secret{}

	err = client.Get(ctx, types.NamespacedName{Name: secret, Namespace: namespace}, credSecret)

	if err != nil {
		return nil, errors.Wrap(err, "error during credential secret lookup")
	}

	key, ok := credSecret.Data["PACKET_AUTH_TOKEN"]
	if !ok {
		return nil, fmt.Errorf("no key PACKET_AUTH_TOKEN found in secret %s", secret)
	}

	m.Client = packngo.NewClientWithAuth("packngo lib", string(key), nil)

	projectID, ok := credSecret.Data["PROJECT_ID"]
	if !ok {
		return nil, fmt.Errorf("no key PROJECT_ID specified in secret %s", secret)
	}

	m.ProjectID = string(projectID)
	return m, nil
}

// CreateElasticInterface creates and attaches ElasticInterface to instance
func (m *MetalClient) CreateElasticInterface(instance *equinixv1alpha1.Instance) (status *equinixv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	tag := fmt.Sprintf("%s-%s", instance.Name, instance.Namespace)
	tag = "instance-sample-hobbyfarm"
	var found bool
	var reservationID string
	var elasticIP string
	project := m.ProjectID
	if instance.Spec.ProjectID != "" {
		project = instance.Spec.ProjectID
	}
	// find if ip with this name already exists //
	queryParam := make(map[string]string)
	queryParam["tag"] = tag
	reservationList, _, err := m.ProjectIPs.List(project, &packngo.ListOptions{
		QueryParams: queryParam,
	})

	if err != nil && !strings.Contains(err.Error(), "404") {
		return status, err
	}

	if len(reservationList) > 1 {
		return status, fmt.Errorf("multiple elastic interfaces found with the same tag")
	}

	if len(reservationList) == 1 {
		found = true
		reservationID = reservationList[0].ID
		elasticIP = reservationList[0].Address
	}

	// prepare for updates
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}

	if found {
		instance.Annotations[ReservationAnnotation] = reservationID
		instance.Annotations[AddressAnnotation] = elasticIP

	} else {

		ipReq := &packngo.IPReservationRequest{
			Type:     "public_ipv4",
			Quantity: 1,
			Tags:     []string{tag},
			Metro:    &instance.Spec.Metro,
		}

		reservation, _, err := m.Client.ProjectIPs.Request(project, ipReq)
		if err != nil {
			return status, err
		}
		instance.Annotations[ReservationAnnotation] = reservation.ID
		instance.Annotations[AddressAnnotation] = reservation.Address
	}
	status.Status = "elasticipcreated"
	return status, nil
}

func (m *MetalClient) CreateNewDevice(instance *equinixv1alpha1.Instance) (status *equinixv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	dsr := m.generateDeviceCreationRequest(instance)
	device, _, err := m.Devices.Create(dsr)
	if err != nil {
		return status, errors.Wrap(err, "error during device creation")
	}

	status.InstanceID = device.ID
	status.Status = device.State
	return status, err
}

func (m *MetalClient) generateDeviceCreationRequest(instance *equinixv1alpha1.Instance) (dsr *packngo.DeviceCreateRequest) {
	dsr = &packngo.DeviceCreateRequest{
		Hostname:              fmt.Sprintf("%s-%s", instance.Name, instance.Namespace),
		Plan:                  instance.Spec.Plan,
		Facility:              instance.Spec.Facility,
		Metro:                 instance.Spec.Metro,
		ProjectID:             instance.Spec.ProjectID,
		AlwaysPXE:             instance.Spec.AlwaysPXE,
		Tags:                  instance.Spec.Tags,
		Description:           instance.Spec.Description,
		PublicIPv4SubnetSize:  instance.Spec.PublicIPv4SubnetSize,
		HardwareReservationID: instance.Spec.HardwareReservationID,
		SpotInstance:          instance.Spec.SpotInstance,
		SpotPriceMax:          instance.Spec.SpotPriceMax.AsApproximateFloat64(),
		CustomData:            instance.Spec.CustomData,
		UserSSHKeys:           instance.Spec.UserSSHKeys,
		ProjectSSHKeys:        instance.Spec.ProjectSSHKeys,
		Features:              instance.Spec.Features,
		NoSSHKeys:             instance.Spec.NoSSHKeys,
		OS:                    instance.Spec.OS,
		BillingCycle:          instance.Spec.BillingCycle,
		IPXEScriptURL:         instance.Spec.IPXEScriptURL,
	}

	if dsr.ProjectID == "" {
		dsr.ProjectID = m.ProjectID
	}
	return dsr
}

func (m *MetalClient) CheckDeviceStatus(instance *equinixv1alpha1.Instance) (status *equinixv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	deviceStatus, _, err := m.Devices.Get(instance.Status.InstanceID, nil)
	if err != nil {
		return status, err
	}

	if deviceStatus.State == "active" {
		// attach elastic IP //
		_, _, err := m.Client.DeviceIPs.Assign(status.InstanceID, &packngo.AddressStruct{
			Address: instance.Annotations[AddressAnnotation],
		})

		if err != nil {
			return status, err
		}

		status.Status = "active"
		status.PrivateIP = deviceStatus.GetNetworkInfo().PublicIPv4
		status.PublicIP = instance.Annotations[AddressAnnotation]
	}

	return status, nil
}

func (m *MetalClient) DeleteDevice(instance *equinixv1alpha1.Instance) (err error) {

	ok, err := m.deviceExists(instance.Status.InstanceID)
	if err != nil {
		return err
	}

	// device exists. terminate the same.
	if ok {
		_, err = m.Devices.Delete(instance.Status.InstanceID, true)
		if err != nil {
			return err
		}
	}

	// delete elastic interface //
	elasticReservationID, ok := instance.Annotations[ReservationAnnotation]

	if ok {
		_, err = m.ProjectIPs.Remove(elasticReservationID)
		// ignore if IP has already been deleted
		if err != nil && strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	// device doesnt exist.. ignore object
	return nil
}

func (m *MetalClient) deviceExists(instanceID string) (ok bool, err error) {
	devices, _, err := m.Devices.List(m.ProjectID, nil)
	if err != nil {
		return ok, err
	}

	for _, device := range devices {
		if device.ID == instanceID {
			ok = true
			return ok, nil
		}
	}

	return ok, nil

}
