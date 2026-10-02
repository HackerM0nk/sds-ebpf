package identity

import "testing"

func TestExtractContainerID(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "docker cgroup v2",
			contents: "0::/system.slice/docker-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef.scope\n",
			want:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
		{
			name:     "docker cgroup path",
			contents: "12:memory:/docker/abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd\n",
			want:     "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		},
		{
			name:     "containerd cgroup path",
			contents: "0::/kubepods.slice/kubepods-burstable.slice/cri-containerd-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.scope\n",
			want:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:     "non-container process",
			contents: "0::/user.slice/user-1000.slice/session-2.scope\n",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractContainerID(tt.contents); got != tt.want {
				t.Fatalf("extractContainerID() = %q, want %q", got, tt.want)
			}
		})
	}
}
