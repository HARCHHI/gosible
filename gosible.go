package main

import (
	"github.com/HARCHHI/gosible/cmd"
	"log"
)

// func getHostKey(host string) (ssh.PublicKey, error) {
// 	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	var hostKey ssh.PublicKey
// 	for scanner.Scan() {
// 		fields := strings.Split(scanner.Text(), " ")
// 		if len(fields) != 3 {
// 			continue
// 		}
// 		if strings.Contains(fields[0], host) {
// 			var err error
// 			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
// 			if err != nil {
// 				return nil, fmt.Errorf("error parsing %q: %v", fields[2], err)
// 			}
// 			break
// 		}
// 	}

// 	if hostKey == nil {
// 		return nil, fmt.Errorf("no hostkey for %s", host)
// 	}
// 	return hostKey, nil
// }

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
