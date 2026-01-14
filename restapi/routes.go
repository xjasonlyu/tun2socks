package restapi

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/vishvananda/netlink"
)

func init() {
	registerEndpoint("/api/v1/routes", routeRouter())
}

func routeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", listRoutes)
	r.Post("/", addRoute)
	r.Delete("/{cidr}", deleteRoute)
	return r
}

type Route struct {
	CIDR    string `json:"cidr"`
	Gateway string `json:"gateway"`
	Device  string `json:"device"`
	Metric  int    `json:"metric"`
}

func listRoutes(w http.ResponseWriter, r *http.Request) {
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		// If device doesn't exist, return empty list instead of error
		// as routes are attached to the device
		render.JSON(w, r, struct {
			Success bool    `json:"success"`
			Message string  `json:"message"`
			Data    []Route `json:"data"`
		}{
			Success: true,
			Message: "Routes retrieved",
			Data:    []Route{},
		})
		return
	}

	routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError(err.Error()))
		return
	}

	result := []Route{}
	for _, rt := range routes {
		if rt.Dst == nil {
			continue // Skip default route if Dst is nil? Or handle it as 0.0.0.0/0
		}

		gw := ""
		if rt.Gw != nil {
			gw = rt.Gw.String()
		}

		result = append(result, Route{
			CIDR:    rt.Dst.String(),
			Gateway: gw,
			Device:  deviceName,
			Metric:  rt.Priority,
		})
	}

	render.JSON(w, r, struct {
		Success bool    `json:"success"`
		Message string  `json:"message"`
		Data    []Route `json:"data"`
	}{
		Success: true,
		Message: "Routes retrieved",
		Data:    result,
	})
}

func addRoute(w http.ResponseWriter, r *http.Request) {
	var req Route
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	_, dst, err := net.ParseCIDR(req.CIDR)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Invalid CIDR format"))
		return
	}

	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, newError("TUN device not found. Start the device first."))
		return
	}

	route := &netlink.Route{
		Dst:       dst,
		LinkIndex: link.Attrs().Index,
		Priority:  req.Metric,
	}

	if req.Gateway != "" {
		gw := net.ParseIP(req.Gateway)
		if gw == nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, newError("Invalid Gateway IP"))
			return
		}
		route.Gw = gw
	}

	if err := netlink.RouteAdd(route); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError("Failed to add route: "+err.Error()))
		return
	}

	render.JSON(w, r, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    Route  `json:"data"`
	}{
		Success: true,
		Message: "Route added successfully",
		Data:    req,
	})
}

func deleteRoute(w http.ResponseWriter, r *http.Request) {
	encodedCidr := chi.URLParam(r, "cidr")
	cidr, err := url.QueryUnescape(encodedCidr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Invalid CIDR encoding"))
		return
	}

	_, dst, err := net.ParseCIDR(cidr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("Invalid CIDR format"))
		return
	}

	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, newError("TUN device not found"))
		return
	}

	route := &netlink.Route{
		Dst:       dst,
		LinkIndex: link.Attrs().Index,
	}

	if err := netlink.RouteDel(route); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, newError("Failed to delete route: "+err.Error()))
		return
	}

	render.JSON(w, r, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Route deleted successfully",
	})
}
