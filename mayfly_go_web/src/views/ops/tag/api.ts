import Api from '@/common/Api';

export const tagApi = {
    getAccountTags: Api.newGet('/tag-trees/account-has'),
    listByQuery: Api.newGet('/tag-trees/query'),
    getTagTrees: Api.newGet('/tag-trees'),
    saveTagTree: Api.newPost('/tag-trees'),
    delTagTree: Api.newDelete('/tag-trees/{id}'),

    getTeams: Api.newGet('/teams'),
    saveTeam: Api.newPost('/teams'),
    delTeam: Api.newDelete('/teams/{id}'),

    getTeamMem: Api.newGet('/teams/{teamId}/members'),
    saveTeamMem: Api.newPost('/teams/{teamId}/members'),
    delTeamMem: Api.newDelete('/teams/{teamId}/members/{accountId}'),

    getTeamTagIds: Api.newGet('/teams/{teamId}/tags'),
    saveTeamTags: Api.newPost('/teams/{teamId}/tags'),
};
