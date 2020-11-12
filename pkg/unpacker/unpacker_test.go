package unpacker

import (
	"testing"
)

func TestModsUnpacker_getArchivedFilesPathsSizes(t *testing.T) {
	type fields struct {
		rawModsDirName      string
		unpackedWorkDirName string
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]int
		wantErr bool
	}{
		{
			name:    "s-plus",
			fields:  fields{},
			args:    args{dir: "C:/dev/go/src/github.com/d8x/amm/workdir/rawmods"},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ModsUnpacker{
				rawModsDirName:      tt.fields.rawModsDirName,
				unpackedWorkDirName: tt.fields.unpackedWorkDirName,
			}
			got, err := m.getArchivedFilesPathsSizes(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getArchivedFilesPathsSizes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, v := range got {
				t.Logf("file archive: %+v\n", v)
			}
		})
	}
}

func TestModsUnpacker_Unpack(t *testing.T) {

	unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	if err != nil {
		t.Error(err)
	}
	if err := unpacker.Unpack(); err != nil {
		t.Error(err)
	}

}

func TestModsUnpacker_unpackArchive(t *testing.T) {
	unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	if err != nil {
		t.Error(err)
	}
	d, err := unpacker.getFileReader("C:\\dev\\go\\src\\github.com\\d8x\\amm\\workdir\\rawmods\\731604991\\LinuxNoEditor\\Crafting\\Station\\_Model\\DM_CraftingStation.uasset.z")
	if err != nil {
		t.Error(err)
	}

	got, err := unpacker.unpackArchive(d)
	if err != nil {
		t.Error(err)
	}
	t.Logf("got: %s\n", string(got))

}
