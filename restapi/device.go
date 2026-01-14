package restapi

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/vishvananda/netlink"
)

func init() {
	registerEndpoint("/api/v1/device", deviceRouter())
}

func deviceRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getDeviceStatus)
	r.Post("/", createDevice)
	r.Delete("/", deleteDevice)
	return r
}

type DeviceStatus struct {
	Name      string `json:"name"`
	Exists    bool   `json:"exists"`
	Status    string `json:"status"`
	IPAddress string `json:"ipAddress"`
	MTU       int    `json:"mtu"`
}

const (
	deviceName = "tunsocks"
	deviceIP   = "198.18.0.1/15"
)

func getDeviceStatus(w http.ResponseWriter, r *http.Request) {
	status := DeviceStatus{
		Name:   deviceName,
		Status: "down",
	}

	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			// Real error
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError(err.Error()))
			return
		}
		// Not found is fine, Exists=false
	} else {
		status.Exists = true
		status.MTU = link.Attrs().MTU
		if link.Attrs().OperState == netlink.OperUp || link.Attrs().Flags&net.FlagUp != 0 {
			status.Status = "up"
		}

		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err == nil && len(addrs) > 0 {
			status.IPAddress = addrs[0].IPNet.String()
		}
	}

	render.JSON(w, r, struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		Data    DeviceStatus `json:"data"`
	}{
		Success: true,
		Message: "Device status retrieved",
		Data:    status,
	})
}

func createDevice(w http.ResponseWriter, r *http.Request) {
	if _, err := netlink.LinkByName(deviceName); err == nil {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, newError("Device already exists"))
		return
	}

	tun := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name: deviceName,
		},
		Mode: netlink.TUNTAP_MODE_TUN,
	}

	if err := netlink.LinkAdd(tun); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError(err.Error()))
		return
	}

	// Retrieve link to ensure it's created and get index
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError("Failed to retrieve created device"))
		return
	}

	// Set IP
	addr, _ := netlink.ParseAddr(deviceIP)
	if err := netlink.AddrAdd(link, addr); err != nil {
		// Cleanup if failed
		netlink.LinkDel(link)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError("Failed to assign IP: "+err.Error()))
		return
	}

	// Set Up
	if err := netlink.LinkSetUp(link); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError("Failed to bring device up: "+err.Error()))
		return
	}

	// Return status
	getDeviceStatus(w, r)
}

func deleteDevice(w http.ResponseWriter, r *http.Request) {
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, newError("Device not found"))
		return
	}

	if err := netlink.LinkDel(link); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError(err.Error()))
		return
	}

	render.JSON(w, r, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Device removed successfully",
	})
}
