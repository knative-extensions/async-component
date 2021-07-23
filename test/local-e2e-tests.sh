#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export PATH=$PWD:$PATH

dir=$(dirname "${BASH_SOURCE[0]}")
base=$(cd "$dir/.." && pwd)
kn_path=`which kn`
kperf_path=`which kperf`

# find where kn is located
check_for_kn() {
	if [ -z "${KN_PATH}" ]; then
		if [ -x "${kn_path}" ]; then
			echo "✅ Found kn executable: $kn_path"
		else
			echo "🔥 Could not find kn executable, please add it to your PATH or set KN_PATH"
			exit -1
		fi
	else
		echo "✅ KN_PATH is set to: $KN_PATH"
		export PATH=$KN_PATH:$PATH
	fi
}

# find where kperf is located
check_for_kperf() {
	if [ -z "${KPERF_PATH}" ]; then
		if [ -x "${kperf_path}" ]; then
			echo "✅ Found kperf executable: $kperf_path"
		else
			echo "🔥 Could not find kperf executable, please add it to your PATH or set KPERF_PATH"
			exit -1
		fi
	else
		echo "✅ KPERF_PATH is set to: $KPERF_PATH"
		export PATH=$KPERF_PATH:$PATH
	fi
}

# find where ko is located
check_for_ko() {
	if [ -z "${KO_PATH}" ]; then
		if [ -x "${ko_path}" ]; then
			echo "✅ Found ko executable: $ko_path"
		else
			echo "🔥 Could not find ko executable, please add it to your PATH or set KO_PATH"
			exit -1
		fi
	else
		echo "✅ KO_PATH is set to: $KO_PATH"
		export PATH=$KO_PATH:$PATH
	fi
}

# Will create and delete this namespace (used for all tests, modify if you want a different one used)
export KNATIVE_ASYNC_E2E_NAMESPACE=knative_async_e2etests

# Make sure `kn` executable is in path
echo "🔦  Looking for kn"
check_for_kn
echo ""

# Make sure `kperf` executable is in path
echo "🔦  Looking for kperf"
check_for_kperf
echo ""

# Make sure `ko` executable is in path
echo "🔦  Looking for ko"
check_for_ko
echo ""

# Start testing
echo "🧪  Testing"
go test ${base}/test/e2e/ -test.v -tags "e2e ${E2E_TAGS}" "$@"
err=$?

# Output
echo ""
if [ $err -eq 0 ]; then
   echo "✅ Success"
else
	echo "❗️Failure"
fi

exit $err