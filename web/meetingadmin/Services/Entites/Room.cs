using System.Collections.Generic;
using Newtonsoft.Json;

namespace meetingadmin.Services.Meetings.Entities {
    public class CreateRoomRequest {
        [JsonProperty("personNumbers")]
        public IEnumerable<string> PersonNumbers { get; set; }
        [JsonProperty("room")]
        public string Room {get; set; }
    }
}