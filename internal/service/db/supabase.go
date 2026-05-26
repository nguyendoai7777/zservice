package db

import (
	"fmt"

	"github.com/supabase-community/supabase-go"
)

var Supabase *supabase.Client

func InitSupabase(url, key string) {
	var err error
	Supabase, err = supabase.NewClient(url, key, nil)
	if err != nil {
		panic(fmt.Sprintf("Không thể khởi tạo Supabase Client: %v", err))
	}
}
