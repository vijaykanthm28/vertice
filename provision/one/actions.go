/*
** Copyright [2013-2015] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package one

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/megamsys/libgo/cmd"
	"github.com/megamsys/megamd/provision"
	"github.com/megamsys/megamd/provision/one/machine"
)

type runMachineActionsArgs struct {
	box           provision.Box
	writer        io.Writer
	imageId       string
	machineStatus Status
}

type machinesToAdd struct {
	Quantity int
	Status   Status
}

var updateStatusInRiak = action.Action{
	Name: "update-status-riak",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runMachineActionsArgs)
		log.Debugf("update status for machine %s, based on image %s for %s", args.box.GetName(), args.imageID, args.box.Compute)

		mach := machine.Machine{
			Name:          args.box.GetName(),
			ComponentId:   args.box.ComponentId,
			BuildingImage: args.machineStatus,
		}

		mach.SetStatus(mach.BuildingImage)
		return mach, nil
	},
	Backward: func(ctx action.BWContext) {
		c := ctx.FWResult.(machine.Machine)
		args := ctx.Params[0].(runMachineActionsArgs)
		c.SetStatus(provision.StatusError)
	},
}

var createMachine = action.Action{
	Name: "create-machine",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		mach := ctx.Previous.(machine.Machine)
		args := ctx.Params[0].(runMachineActionsArgs)
		log.Debugf("create machine for box %s, based on image %s, with %s", args.box.GetName(), args.imageID, args.box.Compute)
		err := mach.Create(&machine.CreateArgs{
			Name:          mach.Name,
			Image:         args.imageId,
			Compute:       args.Compute,
			BuildingImage: mach.status.String(),
			Provisioner:   args.provisioner,
		})
		if err != nil {
			log.Errorf("error on create container for app %s - %s", args.box.GetName(), err)
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
		c := ctx.FWResult.(machine.Machine)
		args := ctx.Params[0].(runMachineActionsArgs)
		fmt.Fprintf(writer, "\n---- Removing %d old %s ----\n", total, pluralize("unit", total))

		err := c.Remove(args.Provisioner)
		if err != nil {
			log.Errorf("Failed to remove the container %q: %s", c.Name, c.ComponentId, err)
		}
	},
}

var removeOldMachine = action.Action{
	Name: "remove-old-machine",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		mach := ctx.Previous.(machine.Machine)
		args := ctx.Params[0].(runMachineActionArgs)
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		fmt.Fprintf(writer, "\n---- Removing old machine %s ----\n", mach.Name)

		err := mach.Remove(args.provisioner)
		if err != nil {
			log.Errorf("Ignored error trying to remove old machine %q: %s", c.ComponentId, err)
		}
		fmt.Fprintf(writer, " ---> Removed old machine %s [%s]\n", c.ComponentId, c.Name)
		return ctx.Previous, nil
	},
	Backward: func(ctx action.BWContext) {
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var addNewRoutes = action.Action{
	Name: "add-new-routes",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runMachineActionArgs)

		/*routeAlreadyExists, err := getRouteAlreadyAttached(args.imageId)
		if err != nil {
			log.Errorf("[WARNING] cannot get the route name as it already exists: %s", err)
		}*/

		mach := ctx.Previous.(machine.Machine)
		r, err := getRouterForBox(args.box)
		if err != nil {
			return nil, err
		}
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}

		fmt.Fprintf(writer, "\n---- Adding routes to new machine ----\n")
		err = r.AddRoute(mach.Name, mach.Address())
		if err != nil {
			return err
		}
		mach.Routable = true
		fmt.Fprintf(writer, " ---> Added route to machine %s [%s]\n", mach.ComponentId, mach.Name)
		return nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		mach := ctx.FWResult.(machine.Machine)
		r, err := getRouterForBox(args.box)
		if err != nil {
			log.Errorf("[add-new-routes:Backward] Error geting router: %s", err.Error())
		}
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		fmt.Fprintf(w, "\n---- Removing routes from created machine ----\n")
		if mach.Routable {
			err = r.RemoveRoute(mach.Name, mach.Address())
			if err != nil {
				log.Errorf("[add-new-routes:Backward] Error removing route for %s: %s", mach.ComponentId, err.Error())
			}
			fmt.Fprintf(w, " ---> Removed route from unit %s [%s]\n", mach.ComponentId, mach.Name)
		}
	},
	OnError: rollbackNotice,
}

var removeOldRoutes = action.Action{
	Name: "remove-old-routes",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		mach := ctx.FWResult.(machine.Machine)
		r, err := getRouterForBox(args.box)
		if err != nil {
			log.Errorf("[remove-old-routes] Error geting router: %s", err.Error())
			return mach, err
		}
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		fmt.Fprintf(w, "\n---- Removing routes from created machine ----\n")
		if mach.Routable {
			err = r.RemoveRoute(mach.Name, mach.Address())
			if err != nil {
				log.Errorf("[add-new-routes:Backward] Error removing route for %s: %s", mach.ComponentId, err.Error())
			}
			fmt.Fprintf(w, " ---> Removed route from unit %s [%s]\n", mach.ComponentId, mach.Name)
		}
		return mach, nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		mach := ctx.FWResult.(machine.Machine)
		r, err := getRouterForBox(args.box)
		if err != nil {
			log.Errorf("[remove-old-routes:Backward] Error geting router: %s", err.Error())
		}
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		fmt.Fprintf(w, "\n---- Adding back routes to old units ----\n")
		if mach.Routable {

			err = r.AddRoute(mach.Name, mach.Address())
			if err != nil {
				log.Errorf("[remove-old-routes:Backward] Error adding back route for %s: %s", cont.ID, err.Error())
			}
			fmt.Fprintf(w, " ---> Added route to unit %s [%s]\n", cont.ShortID(), cont.ProcessName)
		}
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var followLogs = action.Action{
	Name: "follow-logs",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		c, ok := ctx.Previous.(machine.Machine)
		if !ok {
			return nil, errors.New("Previous result must be a container.")
		}
		args := ctx.Params[0].(runMachineActionsArgs)
		err := c.Logs(args.provisioner, args.writer)
		if err != nil {
			log.Errorf("error on get logs for container %s - %s", c.ID, err)
			return nil, err
		}

		return imageId, nil
	},
	Backward: func(ctx action.BWContext) {
	},
	MinParams: 1,
}

var rollbackNotice = func(ctx action.FWContext, err error) {
	args := ctx.Params[0].(changeUnitsPipelineArgs)
	if args.writer != nil {
		fmt.Fprintf(args.writer, "\n**** ROLLING BACK AFTER FAILURE ****\n ---> %s <---\n", err)
	}
}