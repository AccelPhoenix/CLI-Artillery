package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// ---------- Типы и данные ----------
type Coord int

func (c Coord) String() string { return fmt.Sprintf("%05d", c) } // 5 цифр, как ты хотел

type Snail struct {
	East  int
	North int
}

type Target struct {
	Name  string
	East  Coord
	North Coord
	Alt   int
}

func (t Target) String() string {
	return fmt.Sprintf("%s | East: %s | North: %s | Alt: %d",
		t.Name, t.East, t.North, t.Alt)
}

var snailDeltas = [9]Snail{
	{East: 17, North: 83},
	{East: 50, North: 83},
	{East: 83, North: 83},
	{East: 83, North: 50},
	{East: 83, North: 17},
	{East: 50, North: 17},
	{East: 17, North: 17},
	{East: 17, North: 50},
	{East: 50, North: 50},
}

var lastID int = -1 // -1 означает, что ещё не добавлено
var favID int = -1

var targets = make(map[int]Target, 100)
var globalId int = 1 // начинаем с 1, чтобы ID были >=1

// ---------- Вспомогательные функции ----------
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Возвращает строку с информацией о цели по ID или "Нет", если не найдена
func targetInfo(id int) string {
	if t, ok := targets[id]; ok {
		return fmt.Sprintf("%02d: %s", id, t)
	}
	return "Нет"
}

// ---------- Работа с координатами ----------
func correction(East, North Coord, Snail int) (Coord, Coord) {
	Snail -= 1 // приводим к индексу 0..8
	East *= 100
	East += Coord(snailDeltas[Snail].East)

	North *= 100
	North += Coord(snailDeltas[Snail].North)

	return East, North
}

func createNewTarget(Name string, East, North Coord, Alt, Snail int) (Target, error) {
	var obj Target
	if Snail == 0 {
		obj = Target{
			Name:  Name,
			East:  East,
			North: North,
			Alt:   Alt,
		}
	} else if Snail > 9 || Snail < 0 {
		return obj, errors.New("некорректная улитка (1-9)")
	} else {
		East, North = correction(East, North, Snail)
		obj = Target{
			Name:  Name,
			East:  East,
			North: North,
			Alt:   Alt,
		}
	}
	return obj, nil
}

func editTarget(obj *Target, Name string, East, North Coord, Alt, Snail int) error {
	if Snail > 9 || Snail < 0 {
		return errors.New("некорректная улитка (0-9)")
	}

	if Name != "" {
		obj.Name = Name
	}
	if Alt >= 0 { // предполагаем, что Alt неотрицательный, иначе флаг пропуска
		obj.Alt = Alt
	}

	if Snail != 0 {
		East, North = correction(East, North, Snail)
	}
	// Всегда обновляем координаты, даже если Snail=0, так как пользователь мог изменить East/North
	obj.East = East
	obj.North = North
	return nil
}

// ---------- Хранилище ----------
func addTarget(tgt Target) int {
	id := globalId
	targets[id] = tgt
	globalId++
	return id
}

func removeTarget(id int) {
	delete(targets, id)
}

func clearIdandMap() {
	globalId = 1
	targets = make(map[int]Target, 100)
}

func lenTargets() int {
	return len(targets)
}

func listTargets() string {
	if len(targets) == 0 {
		return "Список целей пуст.\n"
	}
	keys := make([]int, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var sb strings.Builder
	for _, id := range keys {
		sb.WriteString(fmt.Sprintf("%02d: %s\n", id, targets[id]))
	}
	return sb.String()
}

// ---------- Ввод данных ----------

func inputCoords(prompt string, scale int) (Coord, Coord, error) {
	fmt.Printf("%s (%d цифр):", prompt, scale)
	var input string
	fmt.Scanln(&input)
	if len(input) != scale {
		return 0, 0, fmt.Errorf("нужно ровно %d цифр", scale)
	}
	half := scale / 2
	eastStr := input[:half]
	northStr := input[half:]

	eastVal, err := strconv.Atoi(eastStr)
	if err != nil {
		return 0, 0, errors.New("некорректная координата East")
	}
	northVal, err := strconv.Atoi(northStr)
	if err != nil {
		return 0, 0, errors.New("некорректная координата North")
	}
	return Coord(eastVal), Coord(northVal), nil
}

// Чтение строки с пробелами
func inputLine(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// ---------- Меню ----------
func menuAddTarget(snailMode bool, scale int) (int, error) {
	name := inputLine("Название цели: ")
	if name == "" {
		name = fmt.Sprintf("default name: %d", globalId)
	}

	east, north, err := inputCoords("Введите координаты одной строкой", scale)
	if err != nil {
		return 0, err
	}

	var alt int
	fmt.Print("Введите Alt (высоту): ")
	_, err = fmt.Scanln(&alt)
	if err != nil {
		return 0, errors.New("некорректная высота")
	}

	var snail int
	if snailMode {
		fmt.Print("Введите улитку (1-9): ")
		_, err = fmt.Scanln(&snail)
		if err != nil {
			return 0, errors.New("некорректная улитка")
		}
	} else {
		snail = 0
	}

	tgt, err := createNewTarget(name, east, north, alt, snail)
	if err != nil {
		return 0, err
	}
	id := addTarget(tgt)
	return id, nil
}

func submenuTarget(id int) error {
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка ввода:", err)
		return err
	}

	for {
		// Показываем информацию о цели
		tgt, _ := targets[id] // мы знаем, что цель существует
		fmt.Printf("Цель %02d: %s\n", id, tgt)
		fmt.Println("e. Редактировать")
		fmt.Println("d. Удалить")
		fmt.Println("f. Сделать избранной")
		fmt.Println("0. Назад")
		fmt.Print("Выберите действие: ")

		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "e", "E":
			err := menuEditTarget(id)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Println("Цель обновлена.")
			}
			return nil // после редактирования можно сразу выйти из подменю
		case "d", "D":
			removeTarget(id)
			fmt.Printf("Цель %02d удалена.\n", id)
			// сброс глобальных переменных
			if lastID == id {
				lastID = -1
			}
			if favID == id {
				favID = -1
			}
			return nil // возвращаемся, потому что цель удалена
		case "f", "F":
			favID = id
			fmt.Printf("Цель %02d теперь избранная.\n", id)
			return nil
		case "0":
			return nil
		default:
			fmt.Println("Неверная команда.")
		}
		fmt.Println("Нажмите Enter для продолжения...")
		scanner.Scan()
	}
}

func menuEditTarget(id int) error {
	tgt, ok := targets[id]
	if !ok {
		return errors.New("цель не найдена")
	}
	fmt.Printf("Редактирование цели %02d. Оставьте поле пустым, чтобы не менять.\n", id)

	name := inputLine(fmt.Sprintf("Имя [%s]: ", tgt.Name))
	// Обработка координат
	eastStr := inputLine(fmt.Sprintf("East [%s]: ", tgt.East))
	var east Coord = tgt.East
	if eastStr != "" {
		val, err := strconv.Atoi(eastStr)
		if err != nil {
			return errors.New("некорректное значение East")
		}
		east = Coord(val)
	}

	northStr := inputLine(fmt.Sprintf("North [%s]: ", tgt.North))
	var north Coord = tgt.North
	if northStr != "" {
		val, err := strconv.Atoi(northStr)
		if err != nil {
			return errors.New("некорректное значение North")
		}
		north = Coord(val)
	}

	altStr := inputLine(fmt.Sprintf("Alt [%d]: ", tgt.Alt))
	var alt int = tgt.Alt
	if altStr != "" {
		val, err := strconv.Atoi(altStr)
		if err != nil {
			return errors.New("некорректное значение Alt")
		}
		alt = val
	}

	snailStr := inputLine("Улитка (0-9) [текущая неизвестна, введите 0 если не используется]: ")
	var snail int
	if snailStr != "" {
		val, err := strconv.Atoi(snailStr)
		if err != nil {
			return errors.New("некорректное значение улитки")
		}
		snail = val
	}

	// Важно: если улитка не используется (snail=0), координаты East/North могут быть новыми – editTarget это учтёт.
	return editTarget(&tgt, name, east, north, alt, snail)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		clearScreen()

		fmt.Println("=== Артиллерийский CLI ===")
		fmt.Printf("Избранная цель: %s\n", targetInfo(favID))
		fmt.Printf("Последняя добавленная: %s\n", targetInfo(lastID))
		fmt.Printf("Всего целей: %d\n\n", lenTargets())

		fmt.Println("Меню:")
		fmt.Println("1. Добавить цель (по улитке)")
		fmt.Println("2. Добавить цель (по точным координатам)")
		fmt.Println("3. Показать список целей и редактировать")
		fmt.Println("999. Очистить все цели")
		fmt.Println("0. Выход")
		fmt.Print("Выберите действие: ")

		scanner.Scan()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Ошибка чтения ввода:", err)
			return
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			id, err := menuAddTarget(true, 6)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				lastID = id
				fmt.Printf("Цель добавлена, ID: %02d\n", id)
			}

		case "2":
			id, err := menuAddTarget(false, 10)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				lastID = id
				fmt.Printf("Цель добавлена, ID: %02d\n", id)
			}

		case "3":
			clearScreen()
			for {
				if lenTargets() == 0 {
					fmt.Println("Список пуст.")
					break
				}
				fmt.Print(listTargets())
				fmt.Print("Введите ID цели (0 для возврата): ")
				scanner.Scan()
				idStr := strings.TrimSpace(scanner.Text())
				if idStr == "0" {
					break
				}
				id, err := strconv.Atoi(idStr)
				if err != nil || id <= 0 {
					fmt.Println("Некорректный ID.")
					fmt.Println("Нажмите Enter для продолжения...")
					scanner.Scan()
					clearScreen()
					continue
				}
				if _, ok := targets[id]; !ok {
					fmt.Println("Цель не найдена.")
					fmt.Println("Нажмите Enter для продолжения...")
					scanner.Scan()
					clearScreen()
					continue
				}
				// Вызываем подменю
				clearScreen()
				submenuTarget(id)
				// После выхода из подменю обновляем список (clearScreen уже вызывается в начале следующей итерации)
				clearScreen()
			}

		case "999":
			clearIdandMap()
			lastID = -1
			favID = -1
			fmt.Println("Все цели удалены.")

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неизвестная команда.")
			fmt.Println("Нажмите Enter для продолжения...")
			scanner.Scan()
		}
	}
}
