package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Đăng nhập vào MangaHub",
	Long:  `Đăng nhập bằng username/email và password để sử dụng các tính năng của CLI`,
	Run:   runLogin,
}

var (
	apiURL string
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&apiURL, "api-url", "http://localhost:8080", "URL của API server")
}

func runLogin(cmd *cobra.Command, args []string) {
	// Nhập username/email
	fmt.Print("Username hoặc Email: ")
	var usernameOrEmail string
	fmt.Scanln(&usernameOrEmail)

	if usernameOrEmail == "" {
		fmt.Println("❌ Username/Email không được để trống")
		os.Exit(1)
	}

	// Nhập password (ẩn)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Xuống dòng sau khi nhập password
	if err != nil {
		fmt.Printf("❌ Lỗi đọc password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)

	if password == "" {
		fmt.Println("❌ Password không được để trống")
		os.Exit(1)
	}

	fmt.Println("\n🔄 Đang đăng nhập...")

	// Tạo HTTP client
	httpClient := client.NewHTTPClient(apiURL)

	// Đăng nhập
	resp, err := httpClient.Login(usernameOrEmail, password)
	if err != nil {
		fmt.Printf("❌ Đăng nhập thất bại: %v\n", err)
		os.Exit(1)
	}

	// Parse user info
	userMap, ok := resp.User.(map[string]interface{})
	if !ok {
		fmt.Println("❌ Không thể parse thông tin user")
		os.Exit(1)
	}

	userID, _ := userMap["id"].(float64)
	username, _ := userMap["username"].(string)

	// Lưu config
	cfg := &config.Config{
		Token:    resp.Token,
		UserID:   int64(userID),
		Username: username,
		APIURL:   apiURL,
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("⚠ Cảnh báo: Không thể lưu config: %v\n", err)
	}

	fmt.Println("\n✅ Đăng nhập thành công!")
	fmt.Printf("👤 Chào mừng, %s (ID: %d)\n", username, int64(userID))
	fmt.Println("\n💡 Bạn có thể sử dụng các lệnh sau:")
	fmt.Println("   • mangahub list-manga        - Xem danh sách manga")
	fmt.Println("   • mangahub show-manga <id>   - Xem chi tiết manga")
	fmt.Println("   • mangahub read-chapter      - Đọc chapter")
	fmt.Println("   • mangahub sync-progress     - Đồng bộ tiến độ (TCP)")
	fmt.Println("   • mangahub notifications     - Nhận thông báo (UDP)")
}
