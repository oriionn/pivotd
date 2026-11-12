package main

/*
#cgo pkg-config: ddcutil
#include <ddcutil_c_api.h>

static DDCA_Display_Info* display_at(DDCA_Display_Info_List* l, int i) {
	return &l->info[i];
}

static int display_busno(DDCA_Display_Info* d) {
	return d->path.io_mode == DDCA_IO_I2C ? d->path.path.i2c_busno : -1;
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.oriondev.fr/orion/pivotd/src/common"
)

type Monitor struct {
	Index     int
	ModelName string
	Name      string
	CurrentRotation common.Rotation

	dref C.DDCA_Display_Ref
}

func drmConnector(d *C.DDCA_Display_Info) (string, bool) {
	var info *C.DDCA_Display_Info2
	if rc := C.ddca_get_display_info2(d.dref, &info); rc != 0 {
		return "", false
	}
	defer C.ddca_free_display_info2(info)

	if info.drm_card_connector_found_by == C.DDCA_DRM_CONNECTOR_NOT_FOUND {
		return "", false
	}
	full := C.GoString(&info.drm_card_connector[0])
	_, name, ok := strings.Cut(full, "-")
	return name, ok && name != ""
}

func ConnectorName(busno int) (string, error) {
	want := fmt.Sprintf("i2c-%d", busno)

	links, _ := filepath.Glob("/sys/class/drm/card*-*/ddc")
	for _, l := range links {
		target, err := os.Readlink(l)
		if err != nil || filepath.Base(target) != want {
			continue
		}
		dir := filepath.Base(filepath.Dir(l))
		if _, name, ok := strings.Cut(dir, "-"); ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("no DRM connector for i2c-%d", busno)
}

func RedetectMonitors() error {
	if rc := C.ddca_redetect_displays(); rc != 0 {
		return fmt.Errorf("ddca redetect error %d", int(rc))
	}
	return nil
}

func GetMonitors() ([]Monitor, error) {
	var list *C.DDCA_Display_Info_List
	if rc := C.ddca_get_display_info_list2(C.bool(false), &list); rc != 0 {
		return nil, fmt.Errorf("ddca error %d", int(rc));
	}
	defer C.ddca_free_display_info_list(list)

	var monitors []Monitor
	for i := 0; i < int(list.ct); i++ {
		d := C.display_at(list, C.int(i))

		// Skip displays without a readable orientation sensor
		if _, err := readOrientation(d.dref); err != nil { continue }

		var err error
		connector, ok := drmConnector(d)
		if !ok {
			connector, err = ConnectorName(int(C.display_busno(d)))
			if err != nil { continue }
		}

		monitors = append(monitors, Monitor{
			Index:     i,
			ModelName: C.GoString(&d.model_name[0]),
			Name:      connector,
			// Unknown on purpose: the compositor state may not match the
			// physical orientation, so the first reading must always be applied.
			CurrentRotation: 0,
			dref:            d.dref,
		})
	}

	return monitors, nil;
}

func (m *Monitor) GetOrientation() (common.Rotation, error) {
	return readOrientation(m.dref)
}

func readOrientation(dref C.DDCA_Display_Ref) (common.Rotation, error) {
	var dh C.DDCA_Display_Handle
	if rc := C.ddca_open_display2(dref, C.bool(false), &dh); rc != 0 {
		return -1, fmt.Errorf("open error %d", int(rc))
	}
	defer C.ddca_close_display(dh)

	var val C.DDCA_Non_Table_Vcp_Value
	if rc := C.ddca_get_non_table_vcp_value(dh, 0xAA, &val); rc != 0 {
		return -1, fmt.Errorf("unreadable orientation")
	}

	switch int(val.sl) {
	case 2:
		return common.RotationLeft, nil
	case 3:
		return common.RotationInverted, nil
	case 4:
		return common.RotationRight, nil
	default:
		return common.RotationNone, nil
	}
}
