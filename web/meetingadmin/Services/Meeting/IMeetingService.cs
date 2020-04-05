using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using meetingadmin.Domain;
using meetingadmin.Services.Meetings.Entities;
using Newtonsoft.Json;


namespace meetingadmin.Services.Meetings
{
    public interface IMeetingService
    {
        Task<bool> AddMeeting(Meeting meeting);
    }

    public class MeetingService : IMeetingService
    {
        private static readonly HttpClient httpClient;

        static MeetingService() {
            httpClient = new HttpClient() {
                BaseAddress = new Uri(System.Environment.GetEnvironmentVariable("MDB_URL") ?? "http://localhost:8081/")
            };
        }

        public async Task<bool> AddMeeting(Meeting meeting)
        {
            var serialized = JsonConvert.SerializeObject(MapRoom(meeting));

            var content = new StringContent(serialized);

            var res = await httpClient.PostAsync("create", content);

            return res.IsSuccessStatusCode;
        }

        private Entities.CreateRoomRequest MapRoom(Meeting meeting) {
            return new Entities.CreateRoomRequest() {
                Room = meeting.RoomName,
                PersonNumbers = meeting.PersonNumbers
            };
        }
    }
}