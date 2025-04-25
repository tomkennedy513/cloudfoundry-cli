package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

type BuildpacksCommand struct {
	BaseCommand

	// todo: add lifecycle to usage
	usage           interface{} `usage:"CF_NAME buildpacks [--labels SELECTOR]\n\nEXAMPLES:\n   CF_NAME buildpacks\n   CF_NAME buildpacks --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME buildpacks --labels 'env=dev,!chargeback-code,tier in (backend,worker)'"`
	relatedCommands interface{} `related_commands:"create-buildpack, delete-buildpack, rename-buildpack, update-buildpack"`
	Labels          string      `long:"labels" description:"Selector to filter buildpacks by labels"`
	Lifecycle       string      `long:"lifecycle" description:"Filter buildpacks with by lifecycle ('buildpack' or 'cnb')"`
}

func (cmd BuildpacksCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting buildpacks as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	// buildpacks, warnings, err := cmd.Actor.GetBuildpacks(cmd.Labels, cmd.Lifecycle)
	// cmd.UI.DisplayWarnings(warnings)
	// if err != nil {
	// 	return err
	// }
	buildpacks, warnings, err := cmd.Actor.GetBuildpacks(cmd.Labels, "cnb")
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	buildpacks2, warnings2, err := cmd.Actor.GetBuildpacks(cmd.Labels, "buildpack")
	cmd.UI.DisplayWarnings(warnings2)
	if err != nil {
		return err
	}
	buildpacks = append(buildpacks, buildpacks2...)

	if len(buildpacks) == 0 {
		cmd.UI.DisplayTextWithFlavor("No buildpacks found")
	} else {
		cmd.displayTable(buildpacks)
	}
	return nil
}

func (cmd BuildpacksCommand) displayTable(buildpacks []resources.Buildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "state", "filename", "lifecycle"},
		}

		// todo: can we always rely on it beign sorted by capi, also do i need to account for position being unset
		// im assuming i dont need this sicne we arent sorting today
		// sort.Slice(buildpacks, func(i, j int) bool {
		// 	if buildpacks[i].Lifecycle != buildpacks[j].Lifecycle {
		// 		return buildpacks[i].Lifecycle < buildpacks[j].Lifecycle
		// 	}
		//
		// 	return buildpacks[i].Position.Value < buildpacks[j].Position.Value
		// })

		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{
				strconv.Itoa(buildpack.Position.Value),
				buildpack.Name,
				buildpack.Stack,
				strconv.FormatBool(buildpack.Enabled.Value),
				strconv.FormatBool(buildpack.Locked.Value),
				buildpack.State,
				buildpack.Filename,
				buildpack.Lifecycle,
			})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}

func (cmd BuildpacksCommand) displayTablesByLifecycle(buildpacks []resources.Buildpack) {
	if len(buildpacks) == 0 {
		return
	}

	lifecycleMap := groupByLifecycle(buildpacks)
	for lifecycleName, buildpacksSlice := range lifecycleMap {
		if lifecycleName != "" {
			cmd.UI.DisplayHeader(fmt.Sprintf("\nlifecycle: %s", lifecycleName))
		}

		cmd.displayTable(buildpacksSlice)
	}
}

func groupByLifecycle(buildpacks []resources.Buildpack) map[string][]resources.Buildpack {
	lifecycleMap := map[string][]resources.Buildpack{}

	for _, buildpack := range buildpacks {
		arr := lifecycleMap[buildpack.Lifecycle]
		if arr == nil {
			arr = make([]resources.Buildpack, 0)
		}
		arr = append(arr, buildpack)
		lifecycleMap[buildpack.Lifecycle] = arr
	}

	return lifecycleMap

}
