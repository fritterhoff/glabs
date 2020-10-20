package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type AssignmentConfig struct {
	Course            string
	Name              string
	Path              string
	Per               Per
	Description       string
	ContainerRegistry bool
	AccessLevel       AccessLevel
	Students          []string
	Groups            []*Group
	Startercode       *Startercode
}

type Per string

const (
	PerStudent Per = "Student"
	PerGroup   Per = "Group"
	PerFailed  Per = "could not happen"
)

type Startercode struct {
	URL             string
	FromBranch      string
	ToBranch        string
	ProtectToBranch bool
}

type Group struct {
	GroupName string
	Members   []string
}

type AccessLevel int

const (
	Guest      AccessLevel = 10
	Reporter   AccessLevel = 20
	Developer  AccessLevel = 30
	Maintainer AccessLevel = 40
)

func GetAssignmentConfig(course, assignment string, onlyForStudentsOrGroups ...string) *AssignmentConfig {
	if !viper.IsSet(course) {
		log.Fatal().
			Str("course", course).
			Msg("configuration for course not found")
	}

	if !viper.IsSet(course + "." + assignment) {
		log.Fatal().
			Str("course", course).
			Str("assignment", assignment).
			Msg("configuration for assignment not found")
	}

	assignmentKey := course + "." + assignment
	per := per(assignmentKey)

	assignmentConfig := &AssignmentConfig{
		Course:            course,
		Name:              assignment,
		Path:              assignmentPath(course, assignment),
		Per:               per,
		Description:       description(assignmentKey),
		ContainerRegistry: viper.GetBool(assignmentKey + ".containerRegistry"),
		AccessLevel:       accessLevel(assignmentKey),
		Students:          students(per, course, onlyForStudentsOrGroups...),
		Groups:            groups(per, course, onlyForStudentsOrGroups...),
		Startercode:       startercode(assignmentKey),
	}

	return assignmentConfig
}

func assignmentPath(course, assignment string) string {
	path := viper.GetString(course + ".coursepath")
	if semesterpath := viper.GetString(course + ".semesterpath"); len(semesterpath) > 0 {
		path += "/" + semesterpath
	}

	assignmentpath := path
	if group := viper.GetString(course + "." + assignment + ".assignmentpath"); len(group) > 0 {
		assignmentpath += "/" + group
	}

	return assignmentpath
}

func per(assignmentKey string) Per {
	if per := viper.GetString(assignmentKey + ".per"); per == "group" {
		return PerGroup
	}
	return PerStudent
}

func description(assignmentKey string) string {
	description := "generated by glabs"

	if desc := viper.GetString(assignmentKey + ".description"); desc != "" {
		description = desc
	}

	return description
}

func accessLevel(assignmentKey string) AccessLevel {
	accesslevelIdentifier := viper.GetString(assignmentKey + ".accesslevel")

	switch accesslevelIdentifier {
	case "guest":
		return Guest
	case "reporter":
		return Reporter
	case "maintainer":
		return Maintainer
	}

	return Developer
}

func students(per Per, course string, onlyForStudentsOrGroups ...string) []string {
	if per == PerGroup {
		return nil
	}
	students := viper.GetStringSlice(course + ".students")
	if len(onlyForStudentsOrGroups) > 0 {
		onlyForStudents := make([]string, 0, len(onlyForStudentsOrGroups))
		for _, onlyStudent := range onlyForStudentsOrGroups {
			for _, student := range students {
				if onlyStudent == student {
					onlyForStudents = append(onlyForStudents, onlyStudent)
				}
			}
		}
		students = onlyForStudents
	}

	return students
}

func groups(per Per, course string, onlyForStudentsOrGroups ...string) []*Group {
	if per == PerStudent {
		return nil
	}

	groupsMap := viper.GetStringMapStringSlice(course + ".groups")
	if len(onlyForStudentsOrGroups) > 0 {
		onlyTheseGroups := make(map[string][]string)
		for _, onlyGroup := range onlyForStudentsOrGroups {
			for groupname, students := range groupsMap {
				if onlyGroup == groupname {
					onlyTheseGroups[groupname] = students
				}
			}
		}
		groupsMap = onlyTheseGroups
	}

	groups := make([]*Group, 0, len(groupsMap))
	for groupname, members := range groupsMap {
		groups = append(groups, &Group{
			GroupName: groupname,
			Members:   members,
		})
	}

	return groups
}

func startercode(assignmentKey string) *Startercode {
	startercodeMap := viper.GetStringMapString(assignmentKey + ".startercode")

	if len(startercodeMap) == 0 {
		log.Debug().Str("assignmemtKey", assignmentKey).Msg("no startercode provided")
		return nil
	}

	url, ok := startercodeMap["url"]
	if !ok {
		log.Fatal().Str("assignmemtKey", assignmentKey).Msg("startercode provided without url")
		return nil
	}

	fromBranch := "master"
	if fB := viper.GetString(assignmentKey + ".startercode.fromBranch"); len(fB) > 0 {
		fromBranch = fB
	}

	toBranch := "master"
	if tB := viper.GetString(assignmentKey + ".startercode.toBranch"); len(tB) > 0 {
		toBranch = tB
	}

	return &Startercode{
		URL:             url,
		FromBranch:      fromBranch,
		ToBranch:        toBranch,
		ProtectToBranch: viper.GetBool(assignmentKey + ".startercode.protectToBranch"),
	}
}
