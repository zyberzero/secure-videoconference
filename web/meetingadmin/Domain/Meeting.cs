using System;
using System.Collections.Generic;

namespace meetingadmin.Domain
{
    public class Meeting
    {
        public string RoomName {get;set;}
        public IEnumerable<string> PersonNumbers {get;set;}
    }
}
