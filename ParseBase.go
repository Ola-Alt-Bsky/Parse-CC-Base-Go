package main

import (
	"bufio"
    "fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
    // Read input from a .txt file
	var file_path string
	fmt.Println("Welcome! You will need to enter in the location of your file.")
	fmt.Print("Enter in the ABSOLUTE file path of the base txt file: ")
	
	input := bufio.NewScanner(os.Stdin)
    if input.Scan() {
        file_path = input.Text()
    }

	if strings.HasPrefix(file_path, `"`) && strings.HasSuffix(file_path, `"`) {
        file_path = strings.Trim(file_path, `"`)
    }

	// Try to retrieve the text from the file
	file, err := os.Open(file_path)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
	defer file.Close()

	var file_lines []string
    reader := bufio.NewScanner(file)
    for reader.Scan() {
		file_lines = append(file_lines, reader.Text())
    }
    if err := reader.Err(); err != nil {
        fmt.Println("Error reading file:", err)
    }

	// Parse and convert to JSON
	parsed_map := parse_to_map(file_lines)
	json_content, char_content, loc_content, song_content := retrieve_from_map(parsed_map)

	// Save the parsed JSON information to a folder
	parentDir := filepath.Dir(file_path)
	outputDir := filepath.Join(parentDir, "Output")
	
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	fileName := filepath.Join(outputDir, "Casual Roleplay.json")
	if err = os.WriteFile(fileName, []byte(json_content), 0666); err != nil {
        fmt.Println(err)
    }
	fmt.Println("Parsed JSON has been saved to Casual Roleplay.json.")

	fileName = filepath.Join(outputDir, "Casual Roleplay Characters.txt")
	if err = os.WriteFile(fileName, []byte(char_content), 0666); err != nil {
        fmt.Println(err)
    }
	fmt.Println("Parsed JSON has been saved to Casual Roleplay Characters.txt.")

	fileName = filepath.Join(outputDir, "Casual Roleplay Locations.txt")
	if err = os.WriteFile(fileName, []byte(loc_content), 0666); err != nil {
        fmt.Println(err)
    }
	fmt.Println("Parsed JSON has been saved to Casual Roleplay Locations.txt.")

	fileName = filepath.Join(outputDir, "Casual Roleplay Songs.txt")
	if err = os.WriteFile(fileName, []byte(song_content), 0666); err != nil {
        fmt.Println(err)
    }
	fmt.Println("Parsed JSON has been saved to Casual Roleplay Songs.txt.")
}

func parse_to_map(file_lines []string) map[string]Season {
	info := make(map[string]Season)

	last_season := "null"
	last_episode := "null"
	last_attribute := "null"
	last_content := "null"
	last_specific := "null"

	for _, line := range file_lines {
		starts_with_star := line[0] == '*'
		starts_with_space := line[0] == ' '
		amount_leading_space := len(line) - len(strings.Trim(line, " "))

		if !(starts_with_star || starts_with_space) { // Season
			line = strings.TrimPrefix(line, "\uFEFF")
			info[line] = Season{episodes: make(map[string]Episode)}
			last_season = line
		} else if starts_with_star { // Episode
			line = strings.Trim(line, "* ")
			info[last_season].episodes[line] = Episode{}
			last_episode = line
		} else if starts_with_space && amount_leading_space == 3 { // Attributes
			line = strings.Trim(line, "* ")
			last_attribute = line
		} else if starts_with_space && amount_leading_space == 6 { // Content
			line = strings.Trim(line, "* ")
			cur_episode := info[last_season].episodes[last_episode]

			switch (last_attribute) {
			case "Characters":
				cur_episode.characters = append(cur_episode.characters, line)
			case "Locations":
				cur_episode.locations = append(cur_episode.locations, line)
			case "Start Date":
				cur_episode.start_date = line
			case "Timeline":
				cur_episode.timeline = line
			case "Songs":
				last_content = line
			default:
				println("Default Case Found")
			}

			info[last_season].episodes[last_episode] = cur_episode
		} else if starts_with_space && amount_leading_space == 9 { // Specific
			line = strings.Trim(line, "* ")
			cur_episode := info[last_season].episodes[last_episode]

			switch (last_content) {
			case "Intro Song":
				cur_episode.songs.intro_song = line
			case "Scene Specific":
				last_specific = line
			case "Outro Song":
				cur_episode.songs.outro_song = line
			}

			info[last_season].episodes[last_episode] = cur_episode
		} else if starts_with_space && amount_leading_space == 12 {
			line = strings.Trim(line, "* ")
			cur_episode := info[last_season].episodes[last_episode]
			cur_songs := cur_episode.songs

			if cur_songs.scene_specific == nil {
				cur_episode.songs.scene_specific = make(map[string]string)
			}
			cur_episode.songs.scene_specific[last_specific] = line


			info[last_season].episodes[last_episode] = cur_episode
		}
	}

	// Remove extra stuff
	delete(info, "Chapter Template")
	delete(info, "Extra Songs")
	
	return info
}

func retrieve_from_map(info map[string]Season) (string, string, string, string) {
	json_string := "{\n"
	character_string := ""
	location_string := ""
	song_string := ""


	indent_len := 4
	num_szn := len(info)
	szn_counter := 0

	for season_name, season := range(info) {
		szn_counter++

		// Get the season name name
		indent_lvl := 1
		season_name_string := strings.Repeat(" ", indent_len * indent_lvl) + "\"" + season_name + "\": {\n"
		json_string = json_string + season_name_string

		num_eps := len(season.episodes)
		ep_counter := 0
		for episode_name, episode := range(season.episodes) {
			ep_counter++

			// Get the episode name
			indent_lvl = 2
			episode_name_string := strings.Repeat(" ", indent_len * indent_lvl) + "\"" + episode_name + "\": {\n"
			json_string = json_string + episode_name_string

			// Get the characters
			new_char_list, new_char_json := get_json_arr_str("Characters", episode.characters, character_string)
			character_string = character_string + new_char_list
			json_string = json_string + new_char_json

			// Get the locations
			new_loc_list, new_loc_json := get_json_arr_str("Locations", episode.locations, location_string)
			location_string = location_string + new_loc_list
			json_string = json_string + new_loc_json

			// Get the time information
			json_string = json_string + get_json_pair_str("Start Date", episode.start_date)
			json_string = json_string + get_json_pair_str("Timeline", episode.timeline)

			// Get Song information
			new_song_list, new_song_json := get_json_song_str(episode.songs, song_string)
			song_string = song_string + new_song_list
			json_string = json_string + new_song_json

			indent_lvl = 2
			if ep_counter < num_eps {
				json_string = json_string + strings.Repeat(" ", indent_len * indent_lvl) + "},\n"
			} else {
				json_string = json_string + strings.Repeat(" ", indent_len * indent_lvl) + "}\n"
			}
		}

		indent_lvl = 1
		if szn_counter < num_szn {
			json_string = json_string + strings.Repeat(" ", indent_len * indent_lvl) + "},\n"
		} else {
			json_string = json_string + strings.Repeat(" ", indent_len * indent_lvl) + "}\n"
		}
	}

	json_string = json_string + "}\n"

	return json_string, character_string, location_string, song_string
}

func get_json_arr_str(arr_name string, arr_obj []string, item_string string) (string, string) {
	indent_len := 4
	indent_lvl := 3
	json_str := strings.Repeat(" ", indent_len * indent_lvl) + "\"" + arr_name + "\": [\n"

	list_str := ""

	indent_lvl = 4
	for i, name := range(arr_obj) {
		if !strings.Contains(item_string, name) {list_str = list_str + name + "\n"}

		item_string := strings.Repeat(" ", indent_len * indent_lvl) + "\"" + name
		if i < len(arr_obj) - 1 {
			json_str = json_str + item_string + "\",\n"
		} else {
			json_str = json_str + item_string + "\"\n"
		}
	}

	indent_lvl = 3
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "],\n"

	return list_str, json_str
}

func get_json_pair_str(key string, value string) string {
	indent_len := 4
	indent_lvl := 3
	pair_str := strings.Repeat(" ", indent_len * indent_lvl) + "\"" + key + "\": [\n"
	
	indent_lvl = 4
	pair_str = pair_str + strings.Repeat(" ", indent_len * indent_lvl) + "\""
	pair_str = pair_str + value + "\"\n"
	
	indent_lvl = 3
	pair_str = pair_str + strings.Repeat(" ", indent_len * indent_lvl) + "],\n"

	return pair_str
}

func get_json_song_str(songs_obj Songs, song_string string) (string, string) {
	indent_len := 4

	list_str := ""

	// Songs title
	indent_lvl := 3
	json_str := strings.Repeat(" ", indent_len * indent_lvl) + "\"Songs\": {\n"

	// Intro song
	indent_lvl = 4
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "\"Intro Song\": \""
	json_str = json_str + songs_obj.intro_song + "\",\n"

	if !strings.Contains(song_string, songs_obj.intro_song) {list_str = list_str + songs_obj.intro_song + "\n"}

	// Scene specific
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "\"Scene Specific\": {\n"

	indent_lvl = 5
	num_scene_songs := len(songs_obj.scene_specific)
	scene_song_counter := 0
	for scene_key, scene_val := range(songs_obj.scene_specific) {
		scene_song_counter++
		json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl)
		if scene_song_counter < num_scene_songs {
			json_str = json_str + "\"" + scene_key + "\": \"" + scene_val + "\",\n"
		} else {
			json_str = json_str + "\"" + scene_key + "\": \"" + scene_val + "\"\n"
		}

		if !strings.Contains(song_string, scene_val) {list_str = list_str + scene_val + "\n"}
	}

	indent_lvl = 4
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "},\n"

	// Outro song
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "\"Outro Song\": \""
	json_str = json_str + songs_obj.outro_song + "\"\n"

	if !strings.Contains(song_string, songs_obj.outro_song) {list_str = list_str + songs_obj.outro_song + "\n"}

	indent_lvl = 3
	json_str = json_str + strings.Repeat(" ", indent_len * indent_lvl) + "}\n"

	return list_str, json_str
}

type Season struct {
	episodes map[string]Episode
}

type Episode struct {
	characters []string 
	locations []string 
	start_date string 
	timeline string 
	songs Songs 
}

type Songs struct {
	intro_song string 
	scene_specific map[string]string 
	outro_song string
}
