package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"unsafe"
)

func assert(condition bool, msg string) {
	// Just to save me from myself
	if !condition {
		panic("Assertion failed!: " + msg)
	}
}

func IntToByteArray(num uint16) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	return arr
}

func ByteArrayToInt(arr []byte) uint16 {
	val := uint16(0)
	size := len(arr)
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val
}

func find_instance(search_buff []byte, bytes []byte) (bool, uint16) {
	if len(search_buff) == 0 || len(search_buff) < len(bytes) {
		return false, 0
	}
	for i := 0; i <= len(search_buff)-len(bytes); i++ {
		is_match := true
		for j := 0; j < len(bytes); j++ {
			if search_buff[i+j] != bytes[j] {
				is_match = false
				break
			}
		}
		if is_match {
			/* DANGER casting down */
			return true, uint16(len(search_buff) - i)
		}
	}
	return false, 0
}

func compress_file(input_file_path string, output_file_path string) {
	file, err := os.Open(input_file_path)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	out_file, err := os.Create(output_file_path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out_file.Close()

	const max_buf_size = 65536        /* largest indexable via uint16 */
	const search_buffers_size = 32768 /* half of above */

	buffer := make([]byte, max_buf_size)
	var search_buffer []byte    // := make([]byte, search_buffers_size)
	var lookahead_buffer []byte //:= make([]byte, search_buffers_size)
	assert(search_buffers_size <= max_buf_size, "lookahead and behind buffers to large and will never be filled!")

	for {
		read_size, err := file.Read(buffer)
		print(".")
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		for i := 0; i < read_size-2; {
			// Fill the search buffer
			if i < search_buffers_size {
				// we cant fill the search buffer so just fill with what we can
				search_buffer = buffer[0:i]
			} else {
				search_buffer = buffer[i-search_buffers_size : i]
			}
			// Fill lookahead buffer
			if read_size-i < search_buffers_size {
				lookahead_buffer = buffer[i:read_size]
			} else {
				lookahead_buffer = buffer[i : i+search_buffers_size]
			}

			last_match := uint16(0)
			size := 0
			for k := 2; k <= len(lookahead_buffer)-1; k++ {
				pattern_to_match := lookahead_buffer[0:k]
				matches, dis_to_left := find_instance(search_buffer, pattern_to_match)
				if matches {
					last_match = dis_to_left
					if dis_to_left > 0 {
						search_buffer = search_buffer[len(search_buffer)-int(dis_to_left):]
					}
					size = k
				} else {
					break
				}
			}
			out_file.Write([]byte(IntToByteArray(last_match)))
			out_file.Write([]byte(IntToByteArray(uint16(size))))
			out_file.Write(buffer[i+size : i+size+1])
			i += size + 1
		}
	}
}

func decompress_file(in_file_path string, out_file_path string) {
	file, err := os.Open(in_file_path)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	out_file, err := os.Create(out_file_path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out_file.Close()

	const max_buf_size = unsafe.Sizeof(uint16(0))*2 + unsafe.Sizeof(byte(0))
	buffer := make([]byte, max_buf_size)
	search_buffer := []byte{}

	for {
		read_size, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		if read_size != int(max_buf_size) {
			break
		}
		shift := ByteArrayToInt(buffer[0:unsafe.Sizeof(uint16(0))])
		read := ByteArrayToInt(buffer[unsafe.Sizeof(uint16(0)) : unsafe.Sizeof(uint16(0))*2])
		char := buffer[max_buf_size-1]
		if read != 0 {
			//fmt.Println(len(search_buffer)-int(shift), "-", len(search_buffer)-int(shift)+int(read)+1)
			new := search_buffer[len(search_buffer)-int(shift) : len(search_buffer)-int(shift)+int(read)]
			search_buffer = append(search_buffer, new...)
			out_file.Write(new)
		}
		search_buffer = append(search_buffer, char)
		out_file.Write([]byte{char})
	}
}

func main() {
	var in_file string
	var out_file string
	var decompress bool

	flag.StringVar(&in_file, "i", "", "Path to the file to be compressed/decompressed")
	flag.StringVar(&out_file, "o", "", "Path to the output file")
	flag.BoolVar(&decompress, "d", false, "Decompress file instead of compress")
	flag.Parse()
	if in_file == "" || out_file == "" {
		flag.Usage()
		return
	}
	if decompress {
		decompress_file(in_file, out_file)
	} else {
		compress_file(in_file, out_file)
	}
}
