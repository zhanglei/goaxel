/*                                                                              
 * Copyright (C) 2013 Deepin, Inc.                                                 
 *               2013 Leslie Zhai <zhaixiang@linuxdeepin.com>                   
 *                                                                              
 * This program is free software: you can redistribute it and/or modify         
 * it under the terms of the GNU General Public License as published by         
 * the Free Software Foundation, either version 3 of the License, or            
 * any later version.                                                           
 *                                                                              
 * This program is distributed in the hope that it will be useful,              
 * but WITHOUT ANY WARRANTY; without even the implied warranty of               
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                
 * GNU General Public License for more details.                                 
 *                                                                              
 * You should have received a copy of the GNU General Public License            
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.        
 */

package conn

import (
    "fmt"
    "os"
    "path"
)

type CONN struct {
    Protocol    string
    Host        string
    Port        int
    UserAgent   string
    UserName    string
    Passwd      string
    Path        string
    Debug       bool
    Callback    func(int)
    http        HTTP
    ftp         FTP
}

func (conn *CONN) httpConnect() {
    conn.http.Debug = conn.Debug
    conn.http.Protocol = conn.Protocol
    conn.http.UserAgent = conn.UserAgent
    conn.http.Connect(conn.Host, conn.Port)
}

func (conn *CONN) ftpConnect() {
    conn.ftp.Debug = conn.Debug
    conn.ftp.Connect(conn.Host, conn.Port)
    if conn.UserName == "" { conn.UserName = "anonymous" }
    conn.ftp.Login(conn.UserName, conn.Passwd)
    if conn.ftp.Code == 530 {
        fmt.Println("ERROR: login failure")
        return
    }
    conn.ftp.Request("TYPE I")
    dir := path.Dir(conn.Path)
    if dir != "/" { dir += "/" }
    conn.ftp.Cwd(dir)
    return
}

func (conn *CONN) GetContentLength(fileName string) (length int, accept bool) {
    length = 0
    accept = false

    if conn.Protocol == "http" || conn.Protocol == "https" {
        conn.httpConnect()
        conn.http.Get(conn.Path, 1, 0)
        conn.http.Response()
        length = conn.http.GetContentLength()
        accept = conn.http.IsAcceptRange()
    } else if conn.Protocol == "ftp" {
        conn.ftpConnect()
        length = conn.ftp.Size(fileName)
        accept = true
    }

    return
}

func (conn *CONN) Get(range_from, range_to int, f *os.File, fileName string) {
    if conn.Protocol == "http" || conn.Protocol == "https" {
        conn.httpConnect()
        conn.http.Callback = conn.Callback
        conn.http.Get(conn.Path, range_from, range_to)
        conn.http.WriteToFile(f)
    } else if conn.Protocol == "ftp" {
        conn.ftpConnect()
        conn.ftp.Callback = conn.Callback
        newPort := conn.ftp.Pasv()
        newConn := conn.ftp.NewConnect(newPort)
        conn.ftp.Request(fmt.Sprintf("REST %d", range_from))
        conn.ftp.Request("RETR " + fileName)
        conn.ftp.WriteToFile(newConn, f)
    }
}
