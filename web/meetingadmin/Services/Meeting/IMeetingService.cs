using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using meetingadmin.Domain;
using meetingadmin.Services.Meetings.Entites;
using Newtonsoft.Json;


namespace meetingadmin.Services.Meetings
{
    public interface IMeetingService
    {
        Task<long> AddMeeting(Meeting meeting);
    }

    public class MeetingService : IMeetingService
    {
        private static readonly HttpClient httpClient;

        static MeetingService() {
            httpClient = new HttpClient() {
                BaseAddress = new Uri(System.Environment.GetEnvironmentVariable("MDB_URL") ?? "http://localhost:8081/")
            };
        }

        public async Task<long> AddMeeting(Meeting meeting)
        {
            var serialized = JsonConvert.SerializeObject(meeting);

            var content = new StringContent(serialized);

            var res = await httpClient.PostAsync("create", content);

            var responseObject = JsonConvert.DeserializeObject<CreateMeetingResponse>(await res.Content.ReadAsStringAsync());

            return responseObject.Meeting;
        }
    }
}