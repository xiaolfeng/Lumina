package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+)\.(0|[1-9][0-9]*))?$`)

type version struct {
	raw      string
	major    int
	minor    int
	patch    int
	channel  string
	sequence int
}

func (v version) isPrerelease() bool {
	return v.channel != ""
}

func parseVersion(value string) (version, error) {
	value = strings.TrimSpace(value)
	matches := versionPattern.FindStringSubmatch(value)
	if matches == nil {
		return version{}, fmt.Errorf(
			"非法版本号 %q；应为 vX.Y.Z 或 vX.Y.Z-通道.N，例如 v1.1.0-beta.20",
			value,
		)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return version{}, fmt.Errorf("解析 major 版本失败: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return version{}, fmt.Errorf("解析 minor 版本失败: %w", err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return version{}, fmt.Errorf("解析 patch 版本失败: %w", err)
	}

	sequence := 0
	if matches[5] != "" {
		sequence, err = strconv.Atoi(matches[5])
		if err != nil {
			return version{}, fmt.Errorf("解析预发布序号失败: %w", err)
		}
	}

	return version{
		raw:      value,
		major:    major,
		minor:    minor,
		patch:    patch,
		channel:  matches[4],
		sequence: sequence,
	}, nil
}

func nextCore(previous, candidate version) bool {
	if candidate.major == previous.major {
		if candidate.minor == previous.minor {
			return candidate.patch == previous.patch+1
		}
		return candidate.minor == previous.minor+1 && candidate.patch == 0
	}
	return candidate.major == previous.major+1 && candidate.minor == 0 && candidate.patch == 0
}

func validateTransition(previous, candidate version) error {
	if previous.isPrerelease() {
		if previous.major == candidate.major && previous.minor == candidate.minor && previous.patch == candidate.patch {
			if !candidate.isPrerelease() {
				return nil
			}
			if candidate.channel != previous.channel {
				return fmt.Errorf(
					"预发布通道不能直接从 %s 切换为 %s；请先发布当前通道的下一序号，或晋升为正式版本",
					previous.channel,
					candidate.channel,
				)
			}
			if candidate.sequence == previous.sequence+1 {
				return nil
			}
			return fmt.Errorf(
				"预发布序号必须连续：%s 的下一版本应为 v%d.%d.%d-%s.%d",
				previous.raw,
				previous.major,
				previous.minor,
				previous.patch,
				previous.channel,
				previous.sequence+1,
			)
		}
		return fmt.Errorf("%s 尚未晋升为正式版本，不能直接发布 %s", previous.raw, candidate.raw)
	}

	if candidate.isPrerelease() {
		if nextCore(previous, candidate) && candidate.sequence == 1 {
			return nil
		}
		return fmt.Errorf(
			"新预发布版本必须基于紧邻的正式版本号且从序号 1 开始：%s 之后不能发布 %s",
			previous.raw,
			candidate.raw,
		)
	}

	if nextCore(previous, candidate) {
		return nil
	}
	return fmt.Errorf("正式版本必须逐级递增：%s 之后不能发布 %s", previous.raw, candidate.raw)
}

func compareVersion(left, right version) int {
	leftCore := []int{left.major, left.minor, left.patch}
	rightCore := []int{right.major, right.minor, right.patch}
	for index := range leftCore {
		if leftCore[index] < rightCore[index] {
			return -1
		}
		if leftCore[index] > rightCore[index] {
			return 1
		}
	}

	if left.isPrerelease() != right.isPrerelease() {
		if left.isPrerelease() {
			return -1
		}
		return 1
	}
	if left.channel < right.channel {
		return -1
	}
	if left.channel > right.channel {
		return 1
	}
	if left.sequence < right.sequence {
		return -1
	}
	if left.sequence > right.sequence {
		return 1
	}
	return 0
}

func findLatest(reader io.Reader, candidate version) (*version, error) {
	var latest *version
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		if tag == "" || tag == candidate.raw {
			continue
		}

		parsed, err := parseVersion(tag)
		if err != nil {
			continue
		}
		if latest == nil || compareVersion(parsed, *latest) > 0 {
			current := parsed
			latest = &current
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取历史版本失败: %w", err)
	}
	return latest, nil
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("check-release-version", flag.ContinueOnError)
	flags.SetOutput(stderr)
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "::error::用法: check-release-version <待发布版本>")
		return 2
	}

	candidate, err := parseVersion(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}

	previous, err := findLatest(stdin, candidate)
	if err != nil {
		fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}
	if previous == nil {
		fmt.Fprintf(stdout, "未找到历史版本，允许首个版本 %s\n", candidate.raw)
		return 0
	}

	if err := validateTransition(*previous, candidate); err != nil {
		fmt.Fprintf(stderr, "::error::版本连续性预检失败：%v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "版本连续性预检通过：%s -> %s\n", previous.raw, candidate.raw)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
