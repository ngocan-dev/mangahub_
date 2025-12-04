@echo off
REM Demo MangaHub CLI

echo ========================================
echo   MangaHub CLI - Demo Test
echo ========================================
echo.

REM Test 1: Show help
echo [Test 1] Hien thi help...
.\mangahub.exe --help
echo.
pause

REM Test 2: Show commands
echo.
echo [Test 2] Cac command co san:
echo   - login           : Dang nhap
echo   - logout          : Dang xuat
echo   - list-manga      : Xem danh sach manga
echo   - show-manga ^<id^> : Xem chi tiet manga
echo   - read-chapter    : Cap nhat tien do doc
echo   - sync-progress   : Dong bo TCP
echo   - notifications   : Nhan thong bao UDP
echo.
pause

REM Test 3: List manga help
echo.
echo [Test 3] List manga options...
.\mangahub.exe list-manga --help
echo.
pause

REM Test 4: TCP sync help
echo.
echo [Test 4] TCP sync options...
.\mangahub.exe sync-progress --help
echo.
pause

REM Test 5: UDP notifications help
echo.
echo [Test 5] UDP notifications options...
.\mangahub.exe notifications --help
echo.
pause

echo.
echo ========================================
echo   Demo hoan thanh!
echo ========================================
echo.
echo De test day du, ban can:
echo 1. Khoi dong backend servers (HTTP, TCP, UDP)
echo 2. Chay: .\mangahub.exe login
echo 3. Test cac command voi server that
echo.
echo Xem chi tiet trong TEST.md
pause
