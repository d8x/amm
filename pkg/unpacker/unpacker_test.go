package unpacker

import (
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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
			m := &ModUnpacker{
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
	// unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	// if err != nil {
	// 	t.Error(err)
	// }
	//
	// if err := unpacker.Unpack(); err != nil {
	// 	t.Error(err)
	// }

	t.Logf("%d", len("d"))

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

func Test_ue4String_Bytes(t *testing.T) {
	type fields struct {
		size int32
		text string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "ModName",
			fields: fields{
				text: "ModName",
			},
			want: []byte{8, 0, 0, 0, 77, 111, 100, 78, 97, 109, 101, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newUE4String(tt.fields.text)
			got := u.Bytes()
			if !reflect.DeepEqual(got, tt.want) {
				t.Error("does not match")
			}
			t.Logf("got: %v", got)
		})
	}
}

func TestModUnpacker_createModFileData(t *testing.T) {
	unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	if err != nil {
		t.Error(err)
	}
	modInfoReader, err := unpacker.getFileReader("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991/mod.info")
	if err != nil {
		t.Error(err)
	}
	modMetaInfoReader, err := unpacker.getFileReader(
		"C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991/modmeta.info")
	if err != nil {
		t.Error(err)
	}
	modInfoData, err := unpacker.unpackModInfo(modInfoReader)
	if err != nil {
		t.Error(err)
	}
	modMetaData, err := unpacker.unpackModMetaInfo(modMetaInfoReader)
	if err != nil {
		t.Error(err)
	}
	modData := unpacker.createModFileData(modInfoData, modMetaData)

	if err := unpacker.writeFile(modData, filepath.Join(unpacker.unpackedWorkDirName,
		strconv.Itoa(int(unpacker.modID))+".mod")); err != nil {
		t.Error(err)
	}
}

func TestModUnpacker_unpackModMetaInfo(t *testing.T) {
	unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	if err != nil {
		t.Error(err)
	}
	modMetaInfoReader, err := unpacker.getFileReader(filepath.Join(unpacker.rawModsDirName, "modmeta.info"))
	if err != nil {
		t.Error(err)
	}
	modeMetaData, err := unpacker.unpackModMetaInfo(modMetaInfoReader)
	if err != nil {
		t.Error(err)
	}
	for _, v := range modeMetaData {
		switch v.key {
		case "ModType":
			assert.Equal(t, 1, v.value)
		case "GameModBaseAsset":
			assert.Equal(t, "TODO", v.value)
		case "PrimalGameData":
			assert.Equal(t, "/Game/Mods/StructuresPlusMod/PrimalGameData_StructuresPlusMod", v.value)
		case "Version":
			assert.Equal(t, "2", v.value)
		case "GUID":
			assert.Equal(t, "E2354DB448F7A3AB7336B6B69379A7B3", v.value)
		}
	}
}

func TestModUnpacker_unpackModInfo(t *testing.T) {
	unpacker, err := NewModsUnpacker("C:/dev/go/src/github.com/d8x/amm/workdir/rawmods/731604991", "amm-unpacked")
	if err != nil {
		t.Error(err)
	}
	modInfoReader, err := unpacker.getFileReader(filepath.Join(unpacker.rawModsDirName, "mod.info"))
	if err != nil {
		t.Error(err)
	}
	modeData, err := unpacker.unpackModInfo(modInfoReader)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int32(18), modeData[0].size)
	assert.Equal(t, "StructuresPlusMod", modeData[0].text)
}
